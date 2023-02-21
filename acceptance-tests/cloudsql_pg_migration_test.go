package acceptance_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/gsql"
	"csbbrokerpakgcp/acceptance-tests/helpers/legacybindings"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

const (
	legacyBrokerName   = "gcp-service-broker"
	legacyDBTier       = "db-f1-micro"
	createExtensionSQL = `
CREATE SCHEMA tiger;
CREATE SCHEMA tiger_data;
CREATE SCHEMA topology;

CREATE EXTENSION IF NOT EXISTS plv8 WITH SCHEMA pg_catalog;
COMMENT ON EXTENSION plv8 IS 'PL/JavaScript (v8) trusted procedural language';
CREATE EXTENSION IF NOT EXISTS address_standardizer WITH SCHEMA public;
COMMENT ON EXTENSION address_standardizer IS 'Used to parse an address into constituent elements. Generally used to support geocoding address normalization step.';
CREATE EXTENSION IF NOT EXISTS fuzzystrmatch WITH SCHEMA public;
COMMENT ON EXTENSION fuzzystrmatch IS 'determine similarities and distance between strings';
CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;
COMMENT ON EXTENSION postgis IS 'PostGIS geometry and geography spatial types and functions';
CREATE EXTENSION IF NOT EXISTS postgis_raster WITH SCHEMA public;
COMMENT ON EXTENSION postgis_raster IS 'PostGIS raster types and functions';
CREATE EXTENSION IF NOT EXISTS postgis_sfcgal WITH SCHEMA public;
COMMENT ON EXTENSION postgis_sfcgal IS 'PostGIS SFCGAL functions';
CREATE EXTENSION IF NOT EXISTS postgis_tiger_geocoder WITH SCHEMA tiger;
COMMENT ON EXTENSION postgis_tiger_geocoder IS 'PostGIS tiger geocoder and reverse geocoder';
CREATE EXTENSION IF NOT EXISTS postgis_topology WITH SCHEMA topology;
COMMENT ON EXTENSION postgis_topology IS 'PostGIS topology spatial types and functions';
`
	createBindingUserGroupSQL = `
create role "binding_user_group" with login;
grant all privileges on all tables in schema public to "binding_user_group";
grant "cloudsqlsuperuser" to "binding_user_group";
do language plpgsql
$$
    declare
        rec record;
    begin
        for rec in
            select usename from pg_catalog.pg_user where usename not like 'cloud%' and usename not in ( 'postgres', 'binding_user_group')
            loop
                execute format('grant "binding_user_group" to %I', rec.usename);
            end loop;
        for rec in
            select datname from pg_catalog.pg_database where datname not in ('cloudsqladmin', 'postgres')
            loop
                execute format('grant all privileges on database %I to "binding_user_group"', rec.datname);
            end loop;
    end
$$
`
	disableBindingUserGroupLoginSQL = `alter role "binding_user_group" with nologin`
)

var _ = Describe("Postgres service instance migration", Label("postgresql-data-migration"), func() {
	It("allows access and reorganisation of migrated data structures", func() {
		By("asynchronously starting the target service instance creation")
		databaseName := random.Name(random.WithPrefix("migrate-database"))
		targetServiceInstance := services.CreateInstance(
			"csb-google-postgres",
			"default",
			services.WithParameters(map[string]any{
				"postgres_version":      "POSTGRES_11",
				"db_name":               databaseName,
				"public_ip":             false,
				"backups_retain_number": 0,
			}),
			services.WithAsync(),
		)

		By("creating the source service instance")
		sourceServiceOffering := "google-cloudsql-postgres-vpc"
		sourceServicePlan := "default"
		sourceServiceInstance := services.CreateInstance(
			sourceServiceOffering,
			sourceServicePlan,
			services.WithBroker(&brokers.Broker{Name: legacyBrokerName}),
			services.WithParameters(map[string]any{
				"tier":            legacyDBTier,
				"private_network": os.Getenv("GCP_PAS_NETWORK"),
				"database_name":   databaseName,
				"backups_enabled": "false",
			}),
		)
		defer sourceServiceInstance.Delete()

		By("binding an app to the source service instance")
		sourceApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer sourceApp.Delete()

		sourceInstanceBinding := sourceServiceInstance.Bind(sourceApp)
		sourceApp.Start()
		defer sourceApp.DELETETestTable()

		By("creating a schema and adding some data in the source database")
		schema := random.Name(random.WithMaxLength(8))
		sourceApp.PUT("", schema)
		defer sourceApp.DELETE(schema)

		key := random.Hexadecimal()
		value := random.Hexadecimal()
		sourceApp.PUT(value, "%s/%s", schema, key)

		By("waiting for the target service creation to succeed")
		services.WaitForInstanceCreation(targetServiceInstance.Name)
		defer targetServiceInstance.Delete()

		By("extracting a service key for the source service instance")
		credentials := sourceInstanceBinding.Credential()
		legacyBinding, err := legacybindings.ExtractPostgresBinding(credentials)
		Expect(err).NotTo(HaveOccurred())

		By("creating a bucket")
		bucketName := "bucket-" + databaseName
		gsql.CreateBackupBucket(bucketName)
		defer gsql.DeleteBucket(bucketName)

		By("creating extensions in the source db")
		gsql.PerformAdminSQL(createExtensionSQL, legacyBinding.InstanceName, legacyBinding.DatabaseName, bucketName)

		By("performing the backup")
		backupURI := gsql.CreateBackup(legacyBinding.InstanceName, databaseName, bucketName)

		By("preparing the restore in the target service instance")
		targetInstanceIaaSName := fmt.Sprintf("csb-postgres-%v", targetServiceInstance.GUID())
		gsql.PerformAdminSQL(createBindingUserGroupSQL, targetInstanceIaaSName, databaseName, bucketName)

		By("restoring the backup onto the target service instance")
		gsql.RestoreBackup(backupURI, targetInstanceIaaSName, databaseName)

		By("cleaning up after the restore")
		gsql.PerformAdminSQL(disableBindingUserGroupLoginSQL, targetInstanceIaaSName, databaseName, bucketName)

		By("binding an app to the target service instance")
		targetApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer targetApp.Delete()
		targetBinding := targetServiceInstance.Bind(targetApp)
		targetApp.Start()

		By("performing the following actions against the target database:")
		By("reading the data")
		got := targetApp.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("creating a new schema")
		newSchema := random.Name(random.WithMaxLength(8))
		newValue := random.Hexadecimal()
		targetApp.PUT("", newSchema)

		By("writing a value")
		newKey := random.Hexadecimal()
		targetApp.PUT(newValue, "%s/%s", newSchema, newKey)

		By("reading the value back")
		gotNewValue := targetApp.GET("%s/%s", newSchema, newKey)
		Expect(gotNewValue).To(Equal(newValue))

		By("modifying the table structure")
		targetApp.PUT("", "schemas/public/test")

		By("unbinding the new user")
		targetBinding.Unbind()
	})
})
