package upgrade_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeSpannerTest", Label("spanner"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-spanner"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			serviceInstance := services.CreateInstance("csb-google-spanner", "small", services.WithBroker(serviceBroker))
			defer serviceInstance.Delete()

			By("pushing the unstarted app")
			appOne := apps.Push(apps.WithApp(apps.Spanner))
			defer apps.Delete(appOne)

			By("binding the app to the service instance")
			binding := serviceInstance.Bind(appOne)

			By("starting the apps")
			apps.Start(appOne)

			By("checking that the app environment has a credhub reference for credentials")
			Expect(binding.Credential()).To(matchers.HaveCredHubRef)

			By("setting a key-value using the app")
			key := random.Hexadecimal()
			value := random.Hexadecimal()
			appOne.PUT(value, key)

			By("getting the value using the same app")
			Expect(appOne.GET(key)).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("checking previously written data still accessible")
			Expect(appOne.GET(key)).To(Equal(value))

			By("updating the instance plan")
			serviceInstance.Update("-p", "medium")

			By("checking previously written data still accessible")
			Expect(appOne.GET(key)).To(Equal(value))
		})
	})
})
