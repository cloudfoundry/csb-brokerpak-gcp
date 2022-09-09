package acceptance_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/gsql"
	"csbbrokerpakgcp/acceptance-tests/helpers/legacybindings"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

var _ = Describe("Postgres service instance migration", func() {
	Expect(true)

	It("retains data", func() {
		By("creating the original service instance")
		sourceServiceInstance := services.CreateInstance("google-cloudsql-postgres", "postgres-db-f1-micro")
		defer sourceServiceInstance.Delete()

		By("binding an app to the source service instance")
		sourceApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer sourceApp.Delete()
		sourceInstanceBinding := sourceServiceInstance.Bind(sourceApp)
		sourceApp.Start()

		By("creating a new service instance with the same version and database name as the original instance")
		legacyBinding, err := legacybindings.ExtractPostgresBinding(sourceInstanceBinding.Credential())
		Expect(err).NotTo(HaveOccurred())

		targetServiceInstance := services.CreateInstance("csb-google-postgres", "default",
			services.WithParameters(map[string]any{"postgres_version": "POSTGRES_11", "db_name": legacyBinding.DatabaseName}))
		defer targetServiceInstance.Delete()

		By("creating a schema and adding some data in the source database")
		schema := random.Name(random.WithMaxLength(8))
		sourceApp.PUT("", schema)

		key := random.Hexadecimal()
		value := random.Hexadecimal()
		sourceApp.PUT(value, "%s/%s", schema, key)

		backupId := gsql.CreateBackup(legacyBinding.InstanceName)

		By("creating a service key for the new service instance")
		serviceKey := targetServiceInstance.CreateServiceKey()
		var serviceKeyMap map[string]interface{}
		serviceKey.Get(&serviceKeyMap)

		gsql.RestoreBackup(fmt.Sprintf("csb-postgres-%v", targetServiceInstance.GUID()), legacyBinding.InstanceName, backupId)

	})
})
