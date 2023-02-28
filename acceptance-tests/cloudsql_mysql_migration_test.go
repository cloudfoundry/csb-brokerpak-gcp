package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/gsql"
	"csbbrokerpakgcp/acceptance-tests/helpers/legacybindings"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MySQL service instance migration", Label("mysql-data-migration"), func() {
	It("allows access to migrated data", func() {
		By("creating a service broker with matching config")
		const mysql56plan = `[{"name":"default","id":"eec62c9b-b25e-4e65-bad5-6b74d90274bf","description":"MySQL v5.6 10GB storage","mysql_version":"MYSQL_5_6","storage_gb":10,"tier":"db-n1-standard-2"}]`

		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mysql"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: brokers.PlansMySQLVar, Value: mysql56plan}),
		)
		defer serviceBroker.Delete()

		By("asynchronously starting the target service instance creation")
		databaseName := random.Name(random.WithPrefix("migrate-database"))
		targetServiceInstance := services.CreateInstance(
			"csb-google-mysql",
			"default",
			services.WithParameters(map[string]any{
				"db_name":                    databaseName,
				"public_ip":                  false,
				"backups_retain_number":      0,
				"allow_insecure_connections": true, // MySQL 5.6 does not support TLS bindings
			}),
			services.WithBroker(serviceBroker),
			services.WithAsync(),
		)

		By("creating the source service instance")
		const sourceServiceOffering = "google-cloudsql-mysql-vpc"
		const sourceServicePlan = "default"
		sourceServiceInstance := services.CreateInstance(
			sourceServiceOffering,
			sourceServicePlan,
			services.WithBrokerName(legacyBrokerName),
			services.WithParameters(map[string]any{
				"tier":            legacyDBTier,
				"private_network": os.Getenv("GCP_PAS_NETWORK"),
				"database_name":   databaseName,
				"backups_enabled": "false",
				"binlog":          "false",
			}),
		)
		defer sourceServiceInstance.Delete()

		By("binding an app to the source service instance")
		sourceApp := apps.Push(apps.WithApp(apps.MySQL))
		defer sourceApp.Delete()

		sourceInstanceBinding := sourceServiceInstance.Bind(sourceApp)
		defer sourceInstanceBinding.Unbind()
		sourceApp.Start()

		By("adding some data in the source database")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		sourceApp.PUT(value, "/key-value/%s", key)

		By("waiting for the target service creation to succeed")
		services.WaitForInstanceCreation(targetServiceInstance.Name)
		defer targetServiceInstance.Delete()

		By("extracting a service key for the source service instance")
		credentials := sourceInstanceBinding.Credential()
		legacyBinding := legacybindings.ExtractLegacyBinding(credentials)

		By("creating a bucket")
		bucketName := "bucket-" + databaseName
		gsql.CreateBackupBucket(bucketName)
		defer gsql.DeleteBucket(bucketName)

		By("performing the backup")
		backupURI := gsql.CreateBackup(legacyBinding.InstanceName, databaseName, bucketName)

		By("preparing the restore in the target service instance")
		targetInstanceIaaSName := fmt.Sprintf("csb-mysql-%s", targetServiceInstance.GUID())

		By("restoring the backup onto the target service instance")
		gsql.RestoreBackup(backupURI, targetInstanceIaaSName, databaseName)

		By("binding an app to the target service instance")
		targetApp := apps.Push(apps.WithApp(apps.MySQL))
		defer targetApp.Delete()
		targetBinding := targetServiceInstance.Bind(targetApp)
		targetApp.Start()

		By("reading the data")
		got := targetApp.GET("/key-value/%s", key)
		Expect(got).To(Equal(value))

		By("writing a new value")
		newKey := random.Hexadecimal()
		newValue := random.Hexadecimal()
		targetApp.PUT(newValue, "/key-value/%s", newKey)

		By("reading the value back")
		gotNewValue := targetApp.GET("/key-value/%s", newKey)
		Expect(gotNewValue).To(Equal(newValue))

		By("unbinding the new user")
		targetBinding.Unbind()
	})
})
