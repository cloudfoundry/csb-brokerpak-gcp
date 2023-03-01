package acceptance_test

import (
	"net"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
)

type pgSSLInfo struct {
	Pid          string `json:"pid"`
	SSL          bool   `json:"ssl"`
	Version      string `json:"version"`
	Cipher       string `json:"cipher"`
	Bits         string `json:"bits"`
	ClientDN     string `json:"clientDN"`
	ClientSerial string `json:"clientSerial"`
	IssuerDN     string `json:"issuerDN"`
}

var _ = Describe("PostgreSQL", Label("postgresql"), func() {
	It("can be accessed by an app", Label("JDBC"), func() {
		By("creating a service broker with Beta services disabled")
		broker := brokers.Create(
			brokers.WithPrefix("csb-postgresql"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: "GSB_COMPATIBILITY_ENABLE_BETA_SERVICES", Value: "false"}),
		)
		defer broker.Delete()

		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-postgres", "small", services.WithBroker(broker))
		defer serviceInstance.Delete()

		By("pushing and starting a first app")
		appOne := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.PostgreSQLTLSTestAppManifest))
		defer apps.Delete(appOne)
		binding := serviceInstance.Bind(appOne)
		apps.Start(appOne)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("setting a key-value using the first app")
		var userIn appResponseUser
		value := random.Hexadecimal()
		appOne.POST("", "?name=%s", value).ParseInto(&userIn)

		By("unbinding the first app to trigger permissions re-assignment")
		binding.Unbind()

		By("pushing and starting a second app")
		appTwo := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.PostgreSQLTLSTestAppManifest))
		defer apps.Delete(appTwo)
		serviceInstance.Bind(appTwo)
		apps.Start(appTwo)

		By("getting the value using the second app")
		var userOut appResponseUser
		appTwo.GET("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		By("verifying the first DB connection utilises TLS")
		var sslInfo pgSSLInfo
		appOne.GET("postgres-ssl").ParseInto(&sslInfo)
		Expect(sslInfo.SSL).To(BeTrue(), "Expected PostgreSQL connection for app %s to be encrypted", appOne.Name)
		Expect(sslInfo.Version).NotTo(HavePrefix("TLSv"))
		Expect(sslInfo.Cipher).NotTo(BeEmpty())
	})

	It("can create instances capable of accepting insecure connection requests", Label("postgresql-no-autotls"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-postgres", "small",
			services.WithParameters(`{"allow_insecure_connections": true}`))
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		appOne := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.PostgreSQLNoAutoTLSTestAppManifest))

		By("binding and starting the app")
		serviceInstance.Bind(appOne)

		appOne.Start()

		By("ensuring encryption wasn't used")
		var sslInfo pgSSLInfo
		appOne.GET("postgres-ssl").ParseInto(&sslInfo)

		Expect(sslInfo.SSL).To(BeFalse(), "Expected PostgreSQL connection for app %s not to be encrypted", appOne.Name)
		Expect(sslInfo.Cipher).To(BeEmpty())
	})

	It("can create service keys with a public IP address", Label("postgresql-public-ip"), func() {
		By("creating a service instance with a public IP address")
		publicIPParams := services.WithParameters(map[string]any{"public_ip": true})
		serviceInstance := services.CreateInstance("csb-google-postgres", "small", publicIPParams)
		defer serviceInstance.Delete()

		By("creating and examining a service key")
		var serviceKeyData struct {
			Credentials struct {
				URI string `json:"uri"`
			} `json:"credentials"`
		}
		serviceKey := serviceInstance.CreateServiceKey()
		serviceKey.Get(&serviceKeyData)

		databaseURI, err := url.ParseRequestURI(serviceKeyData.Credentials.URI)
		Expect(err).NotTo(HaveOccurred())
		uriIP := net.ParseIP(databaseURI.Hostname())
		Expect(uriIP).NotTo(BeNil())
		Expect(uriIP.IsPrivate()).To(BeFalse())
	})
})
