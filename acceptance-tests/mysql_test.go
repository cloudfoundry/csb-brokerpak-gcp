package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type appResponseUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type mySQLOption struct {
	Name  string `json:"variableName"`
	Value string `json:"value"`
}

var _ = Describe("MySQL", Label("mysql"), func() {
	It("can be accessed by an app", Label("JDBC"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-mysql", "default")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.MySQLTLSTestAppManifest))
		appTwo := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.MySQLTLSTestAppManifest))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the storage service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("setting a key-value using the first app")
		value := random.Hexadecimal()
		var userIn appResponseUser
		appOne.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("getting the value using the second app")
		var userOut appResponseUser
		appTwo.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		By("verifying the first DB connection utilises TLS")
		var tlsCipher mySQLOption
		appOne.GET("mysql-ssl").ParseInto(&tlsCipher)

		Expect(strings.ToLower(tlsCipher.Name)).To(Equal("ssl_cipher"))
		Expect(tlsCipher.Value).NotTo(BeEmpty(), "Expected Mysql connection for app %s to be encrypted", appOne.Name)

		By("pushing and binding an app for verifying non-TLS connection attempts")
		golangApp := apps.Push(apps.WithApp(apps.MySQL))
		serviceInstance.Bind(golangApp)
		apps.Start(golangApp)

		By("verifying interactions with TLS enabled")
		key, value := "key", "value"
		golangApp.PUTf(value, "/key-value/%s", key)
		got := golangApp.GETf("/key-value/%s", key).String()
		Expect(got).To(Equal(value))

		By("verifying that non-TLS connections should fail")
		response := golangApp.GETResponsef("/key-value/%s?tls=false", key)
		defer response.Body.Close()
		Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError), "force TLS is enabled by default")
		b, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
		Expect(string(b)).To(ContainSubstring("error connecting to database: failed to verify the connection"), "force TLS is enabled by default")
		Expect(string(b)).To(ContainSubstring("Error 1045 (28000): Access denied for user"), "mysql client cannot connect to the postgres server due to invalid TLS")
	})

	It("can create instances capable of accepting insecure connection requests", Label("mysql-no-autotls"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-mysql", "default",
			services.WithParameters(`{"allow_insecure_connections": true}`))
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		appOne := apps.Push(apps.WithApp(apps.JDBCTestApp), apps.WithTestAppManifest(apps.MySQLNoAutoTLSTestAppManifest))

		By("binding and starting the app")
		serviceInstance.Bind(appOne)

		appOne.Start()

		By("ensuring encryption wasn't used")
		var tlsCipher mySQLOption
		appOne.GETf("mysql-ssl").ParseInto(&tlsCipher)

		Expect(strings.ToLower(tlsCipher.Name)).To(Equal("ssl_cipher"))
		Expect(tlsCipher.Value).To(BeEmpty(), "Expected Mysql connection for app %s not to be encrypted", appOne.Name)
	})

	It("can create service keys with a public IP address", Label("public"), func() {
		By("creating a service instance with a public IP address")
		publicIPParams := services.WithParameters(map[string]any{"public_ip": true})
		serviceInstance := services.CreateInstance("csb-google-mysql", "default", publicIPParams)
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
