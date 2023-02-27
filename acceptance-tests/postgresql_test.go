package acceptance_test

import (
	"encoding/json"
	"io"
	"net"
	"net/url"
	"os"
	"path"

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

var _ = Describe("PostgreSQL", func() {
	It("can be accessed by an app", Label("postgresql"), func() {
		var (
			userIn, userOut appResponseUser
			sslInfo         pgSSLInfo
		)

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

		By("pushing the unstarted app twice")
		testExecutable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())

		testPath := path.Dir(testExecutable)
		appManifest := path.Join(testPath, "apps", "jdbctestapp", "manifest-postgres.yml")
		appOne := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithManifest(appManifest))
		appTwo := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithManifest(appManifest))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("setting a key-value using the first app")
		value := random.Hexadecimal()
		response := appOne.POST("", "?name=%s", value)

		responseBody, err := io.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())

		err = json.Unmarshal(responseBody, &userIn)
		Expect(err).NotTo(HaveOccurred())

		By("getting the value using the second app")
		got := appTwo.GET("%s/%s", schema, key).String()
		Expect(got).To(Equal(value))

		Expect(sslInfo.SSL).To(BeTrue(), "Expected PostgreSQL connection for app %s to be encrypted", appOne.Name)
		Expect(sslInfo.Version).NotTo(HavePrefix("TLSv"))
		Expect(sslInfo.Cipher).NotTo(BeEmpty())
	})

		By("getting the value again using the second app")
		got2 := appTwo.GET("%s/%s", schema, key).String()
		Expect(got2).To(Equal(value))

		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-postgres", "small",
			services.WithParameters(`{"allow_insecure_connections": true}`))
		defer serviceInstance.Delete()

		By("getting the other value using the second app")
		got3 := appTwo.GET("%s/%s", schema, key2).String()
		Expect(got3).To(Equal(value2))

		testPath := path.Dir(testExecutable)
		appManifest := path.Join(testPath, "apps", "jdbctestapp", "manifest-postgres-no-autotls.yml")
		appOne := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithManifest(appManifest))

		By("binding and starting the app")
		serviceInstance.Bind(appOne)

		appOne.Start()

		By("ensuring encryption wasn't used")
		got := appOne.GET("postgres-ssl")
		err = json.Unmarshal([]byte(got), &sslInfo)
		Expect(err).NotTo(HaveOccurred())

		Expect(sslInfo.SSL).To(BeFalse(), "Expected PostgreSQL connection for app %s not to be encrypted", appOne.Name)
		Expect(sslInfo.Cipher).To(BeEmpty())
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
