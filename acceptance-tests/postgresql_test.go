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
			serviceInstance := services.CreateInstance(
				"csb-google-postgres",
				"small",
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

			By("getting the entry using the second app")
			appTwo.GET("%d", userIn.ID).ParseInto(&userOut)
			Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

			By("setting another value using the second app")
			var userInTwo AppResponseUser
			value2 := random.Hexadecimal()
			appOne.POST("", "?name=%s", value2).ParseInto(&userInTwo)

			By("getting the entry using the second app")
			var userOutTwo AppResponseUser
			appTwo.GET("%d", userInTwo.ID).ParseInto(&userOutTwo)
			Expect(userOut.Name).To(Equal(value), "The second app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

			By("deleting the entries using the second app")
			appOne.DELETE("%d", userIn.ID)
			appOne.DELETE("%d", userInTwo.ID)
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
