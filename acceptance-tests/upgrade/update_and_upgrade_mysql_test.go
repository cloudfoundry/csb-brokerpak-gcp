package upgrade_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/plans"
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
				brokers.WithReleasedEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			serviceOffering := "csb-google-mysql"
			servicePlan := "default"
			serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
			// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
			// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
			// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
			// (which can only have 5 instances at any time) to prevent subsequent test failures.
			defer services.Delete(serviceName)
			serviceInstance := services.CreateInstance(
				serviceOffering,
				servicePlan,
				services.WithBroker(serviceBroker),
				services.WithName(serviceName),
			)

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
			appOne.PUTf(value, "/key-value/%s", key)

			By("getting the value using the second app")
			Expect(appTwo.GETf("/key-value/%s", key).String()).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the value using the second app")
			Expect(appTwo.GETf("/key-value/%s", key).String()).To(Equal(value))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("updating the instance plan")
			serviceInstance.Update(services.WithParameters(`{}`))

			By("getting the value using the second app")
			Expect(appTwo.GETf("/key-value/%s", key).String()).To(Equal(value))

			By("deleting bindings created before the update")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("getting the value using the second app")
			Expect(appTwo.GETf("/key-value/%s", key).String()).To(Equal(value))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUTf(valueTwo, "/key-value/%s", keyTwo)
			Expect(appTwo.GETf("/key-value/%s", keyTwo).String()).To(Equal(valueTwo))

			By("verifying the DB connection utilises TLS")
			var sslInfo struct {
				VariableName string `json:"variable_name"`
				Value        string `json:"value"`
			}
			appOne.GETf("/admin/ssl").ParseInto(&sslInfo)

			Expect(sslInfo.VariableName).To(Equal("Ssl_cipher"))
			Expect(sslInfo.Value).To(Equal("ECDHE-RSA-AES128-GCM-SHA256"))
		})
	})
})
