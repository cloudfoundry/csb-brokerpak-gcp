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
	legacyBrokerName = "gcp-service-broker"
	legacyDBTier     = "db-f1-micro"
)

var _ = Describe("Postgres service instance migration", func() {

	FIt("retains data", func() {
		By("asynchronously starting the target service instance creation")
		databaseName := random.Name(random.WithPrefix("migrate-database"))
		targetServiceInstance := services.CreateInstance(
			"csb-google-postgres",
			"default",
			services.WithParameters(map[string]any{
				"postgres_version": "POSTGRES_11",
				"db_name":          databaseName,
				"public_ip":        false,
			}),
			services.WithAsync(),
		)

		By("creating the original service instance")
		sourceServiceOffering := "google-cloudsql-postgres-vpc"
		sourceServicePlan := "default"
		sourceServiceInstance := services.CreateInstance(
			sourceServiceOffering,
			sourceServicePlan,
			services.WithBroker(&brokers.Broker{Name: legacyBrokerName}),
			services.WithParameters(map[string]string{
				"tier":            legacyDBTier,
				"private_network": os.Getenv("GCP_PAS_NETWORK"),
				"database_name":   databaseName,
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

		By("waiting for the new service creation to succeed")
		services.WaitForInstanceCreation(targetServiceInstance.Name)
		defer targetServiceInstance.Delete()

		By("creating a backup for the legacy service instance")
		credentials := sourceInstanceBinding.Credential()
		legacyBinding, err := legacybindings.ExtractPostgresBinding(credentials)
		Expect(err).NotTo(HaveOccurred())

		By("creating a backup bucket")
		bucketName := "bucket-" + databaseName
		gsql.CreateBackupBucket(bucketName)
		defer gsql.DeleteBucket(bucketName)

		backupURI := gsql.CreateBackup(legacyBinding.InstanceName, databaseName, bucketName)

		By("restoring the backup onto the new service instance")
		targetInstanceIaaSName := fmt.Sprintf("csb-postgres-%v", targetServiceInstance.GUID())
		gsql.RestoreBackup(backupURI, targetInstanceIaaSName, databaseName)

		targetApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer targetApp.Delete()
		targetBinding := targetServiceInstance.Bind(targetApp)
		targetApp.Start()

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
