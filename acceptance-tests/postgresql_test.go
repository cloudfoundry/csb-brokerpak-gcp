package acceptance_test

import (
	"net"
	"net/url"

	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

var _ = Describe("PostgreSQL", func() {
	Describe("Can be accessed by an app", func() {
		It("work with JDBC and TLS", Label("JDBC", "postgresql"), func() {
			By("creating a service instance")
			// The test used the small plan.
			// Instances associated with the small plan, that is to say, with db-f1-micro tier,
			// only allow a maximum of 25 connections per instance.
			// See the following link: max_connections => https://cloud.google.com/sql/docs/postgres/flags#postgres-m.
			//
			// We found errors when running the test with the application prepared in Java to test the JDBC URL.
			// The error said the following:
			// failed: unbind could not be completed: Service broker failed to delete service binding for
			// instance csb-google-postgres-small-daffodil-koala: Service broker error: unbind failed:
			// Error: querying for existing role: error finding role "XXXXX": pq: remaining connection
			// slots are reserved for non-replication superuser connections with csbpg_binding_user.new_user,
			// on main.tf ...
			//
			// In a first proof of concept, we intend to change the plan to use an instance that
			// allows more connections. The result is satisfactory. We intend to make a study of the
			// possible causes of the need to use more connections with the Java application than with the
			// Golang application, since the Golang application does not find this limitation.
			// In a first surface analysis of the Terraform provider that creates the bindings,
			// there appears to be no connection leak and the Golang application uses `db.SetMaxIdleConns(0)`.
			//
			// SetMaxIdleConns sets the maximum number of connections in the idle
			// connection pool.
			//
			// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
			// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
			//
			// If n <= 0, no idle connections are retained
			serviceInstance := services.CreateInstance(
				"csb-google-postgres",
				"db-custom-2-7680",
				services.WithParameters(map[string]any{"backups_retain_number": 0}),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.PostgresTestAppManifest))
			appTwo := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.PostgresTestAppManifest))
			defer apps.Delete(appOne, appTwo)

			type AppResponseUser struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}

			type PostgresSSLInfo struct {
				Pid          int    `json:"pid"`
				SSL          bool   `json:"ssl"`
				Version      string `json:"version"`
				Cipher       string `json:"cipher"`
				Bits         int    `json:"bits"`
				ClientDN     string `json:"clientDN"`
				ClientSerial string `json:"clientSerial"`
				IssuerDN     string `json:"issuerDN"`
			}
			var (
				userIn, userOut AppResponseUser
				sslInfo         PostgresSSLInfo
			)

			By("binding the apps to the service instance")
			binding := serviceInstance.Bind(appOne)

			By("starting the first app")
			apps.Start(appOne)

			By("checking that the app environment has a credhub reference for credentials")
			Expect(binding.Credential()).To(matchers.HaveCredHubRef)

			By("creating an entry using the first app")
			value := random.Hexadecimal()
			appOne.POST("", "?name=%s", value).ParseInto(&userIn)

			By("binding and starting the second app")
			serviceInstance.Bind(appTwo)
			apps.Start(appTwo)

			By("getting the entry using the second app")
			appTwo.GET("%d", userIn.ID).ParseInto(&userOut)
			Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

			By("verifying the DB connection utilises TLS")
			appOne.GET("postgres-ssl").ParseInto(&sslInfo)
			Expect(sslInfo.SSL).To(BeTrue())
			Expect(sslInfo.Cipher).NotTo(BeEmpty())
			Expect(sslInfo.Bits).To(BeNumerically(">=", 256))

			By("deleting the entry using the first app")
			appOne.DELETE("%d", userIn.ID)

			By("triggering ownership management")
			binding.Unbind()

			By("setting another value using the second app")
			var userInTwo AppResponseUser
			value2 := random.Hexadecimal()
			appTwo.POST("", "?name=%s", value2).ParseInto(&userInTwo)

			By("getting the entry using the second app")
			var userOutTwo AppResponseUser
			appTwo.GET("%d", userInTwo.ID).ParseInto(&userOutTwo)
			Expect(userOut.Name).To(Equal(value), "The second app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

			By("deleting the entry using the second app")
			appTwo.DELETE("%d", userInTwo.ID)
		})

		It("works with the default postgres version", Label("postgresql"), func() {
			By("creating a service instance")
			serviceInstance := services.CreateInstance("csb-google-postgres", "small")
			defer serviceInstance.Delete()

			postgresTestMultipleApps(serviceInstance)

		})

		It("works with latest changes to public schema in postgres 15", Label("Postgres15"), func() {
			By("creating a service instance")
			serviceInstance := services.CreateInstance("csb-google-postgres", "pg15")
			defer serviceInstance.Delete()

			postgresTestMultipleApps(serviceInstance)

		})

	})

	It("can create service keys with a public IP address", Label("postgresql-public-ip"), func() {
		By("creating a service instance with a public IP address")
		publicIPParams := services.WithParameters(map[string]any{"public_ip": true})
		serviceInstance := services.CreateInstance("csb-google-postgres", "small", publicIPParams)
		defer serviceInstance.Delete()

		By("creating and examining a service key")
		serviceKey := serviceInstance.CreateServiceKey()
		var serviceKeyData map[string]any
		serviceKey.Get(&serviceKeyData)

		Expect(serviceKeyData).To(HaveKey("credentials"))
		creds, _ := serviceKeyData["credentials"].(map[string]any)

		Expect(creds).To(HaveKey("uri"))
		uri, ok := creds["uri"]
		Expect(ok).To(BeTrue())
		uriString, ok := uri.(string)
		Expect(ok).To(BeTrue())
		databaseURI, err := url.ParseRequestURI(uriString)
		Expect(err).NotTo(HaveOccurred())
		uriIP := net.ParseIP(databaseURI.Hostname())
		Expect(uriIP).NotTo(BeNil())
		Expect(uriIP.IsPrivate()).To(BeFalse())
	})
})

func postgresTestMultipleApps(serviceInstance *services.ServiceInstance) {
	GinkgoHelper()

	By("pushing the unstarted app twice")
	appOne := apps.Push(apps.WithApp(apps.PostgreSQL))
	appTwo := apps.Push(apps.WithApp(apps.PostgreSQL))
	defer apps.Delete(appOne, appTwo)

	By("binding the first app to the service instance")
	binding := serviceInstance.Bind(appOne)

	By("starting the first app")
	apps.Start(appOne)

	By("checking that the app environment has a credhub reference for credentials")
	Expect(binding.Credential()).To(matchers.HaveCredHubRef)

	By("creating a schema using the first app")
	schema := random.Name(random.WithMaxLength(10))
	appOne.PUT("", schema)

	By("setting a key-value using the first app")
	key := random.Hexadecimal()
	value := random.Hexadecimal()
	appOne.PUT(value, "%s/%s", schema, key)

	By("binding the second app to the service instance")
	serviceInstance.Bind(appTwo)

	By("starting the second app")
	apps.Start(appTwo)

	By("getting the value using the second app")
	got := appTwo.GET("%s/%s", schema, key).String()
	Expect(got).To(Equal(value))

	By("triggering ownership of schema to pass to provision user")
	binding.Unbind()

	By("getting the value again using the second app")
	got2 := appTwo.GET("%s/%s", schema, key).String()
	Expect(got2).To(Equal(value))

	By("setting another value using the second app")
	key2 := random.Hexadecimal()
	value2 := random.Hexadecimal()
	appTwo.PUT(value2, "%s/%s", schema, key2)

	By("getting the other value using the second app")
	got3 := appTwo.GET("%s/%s", schema, key2).String()
	Expect(got3).To(Equal(value2))

	By("dropping the schema using the second app")
	appTwo.DELETE(schema)

}
