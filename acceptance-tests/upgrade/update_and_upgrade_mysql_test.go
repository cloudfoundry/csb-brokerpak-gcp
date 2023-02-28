package upgrade_test

import (
	"encoding/json"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMYSQLTest", Label("mysql"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-mysql"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleasedEnv(),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			serviceInstance := services.CreateInstance("csb-google-mysql", "small", services.WithBroker(serviceBroker))
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MySQL))
			appTwo := apps.Push(apps.WithApp(apps.MySQL))
			defer apps.Delete(appOne, appTwo)

			By("binding the apps to the storage service instance")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("checking that the app environment has a credhub reference for credentials")
			Expect(bindingOne.Credential()).To(matchers.HaveCredHubRef)

			By("setting a key-value using the first app")
			key := random.Hexadecimal()
			value := random.Hexadecimal()
			appOne.PUT(value, "/key-value/%s", key)

			By("getting the value using the second app")
			Expect(appTwo.GET("/key-value/%s", key)).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			appOne.SetEnv(apps.EnvVar{Name: "NEW_BINDING_FORMAT_FEATURE_FLAG", Value: "ENABLED"})
			appTwo.SetEnv(apps.EnvVar{Name: "NEW_BINDING_FORMAT_FEATURE_FLAG", Value: "ENABLED"})

			By("getting the value using the second app")
			Expect(appTwo.GET("/key-value/%s", key)).To(Equal(value))

			By("updating the instance plan")
			serviceInstance.Update(services.WithPlan("default"))

			By("getting the value using the second app")
			Expect(appTwo.GET("/key-value/%s", key)).To(Equal(value))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("getting the value using the second app")
			Expect(appTwo.GET("/key-value/%s", key)).To(Equal(value))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "/key-value/%s", keyTwo)
			Expect(appTwo.GET("/key-value/%s", keyTwo)).To(Equal(valueTwo))

			By("verifying the DB connection utilises TLS")
			got := appOne.GET("/admin/ssl")
			var sslInfo struct {
				VariableName string `json:"variable_name"`
				Value        string `json:"value"`
			}
			err := json.Unmarshal([]byte(got), &sslInfo)
			Expect(err).NotTo(HaveOccurred())

			Expect("Ssl_cipher").To(Equal(sslInfo.VariableName))
			Expect("ECDHE-RSA-AES128-GCM-SHA256").To(Equal(sslInfo.Value))
		})
	})
})
