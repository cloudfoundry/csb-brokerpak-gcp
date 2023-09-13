package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"net"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

var _ = Describe("PostgreSQL", func() {
	Describe("Can be accessed by an app", func() {
		var broker *brokers.Broker

		BeforeEach(func() {
			broker = brokers.Create(
				brokers.WithPrefix("csb-postgresql"),
				brokers.WithLatestEnv(),
				brokers.WithEnv(apps.EnvVar{Name: "GSB_COMPATIBILITY_ENABLE_BETA_SERVICES", Value: "false"}),
			)
			defer broker.Delete()
		})

		It("works with the default postgres version", Label("postgresql"), func() {
			By("creating a service instance")
			serviceInstance := services.CreateInstance("csb-google-postgres", "small", services.WithBroker(broker))
			defer serviceInstance.Delete()

			postgresTestMultipleApps(serviceInstance)

		})

		It("works with latest changes to public schema in postgres 15", Label("Postgres15"), func() {
			By("creating a service instance")
			serviceInstance := services.CreateInstance("csb-google-postgres", "pg15", services.WithBroker(broker))
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
