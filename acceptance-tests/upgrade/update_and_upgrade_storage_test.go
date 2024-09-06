package upgrade_test

import (
	"fmt"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeStorageTest", Label("storage"), func() {
	FWhen("upgrading the broker to a vm based deployment", func() {
		It("drains in flight instances and waits for them to finish deploying", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-storage"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleasedEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()
			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-google-storage-bucket",
				"default",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("Starting to deploy a vm based broker")
			serviceBrokerVM := brokers.CreateVm(
				brokers.WithName(serviceBroker.Name),
				brokers.WithBoshReleaseDir("../../../csb-gcp-release"),
			)
			defer serviceBrokerVM.Delete()
			By("checking that the vm broker stopped the app broker")
			out, _ := cf.Run("apps")
			Expect(out).To(MatchRegexp(fmt.Sprintf("%s.*stopped", serviceBroker.Name)))

			By("checking that the broker now uses a bosh dns name")
			out, _ = cf.Run("service-brokers")
			Expect(out).To(MatchRegexp(fmt.Sprintf("%s.csb.internal", serviceBrokerVM.Name)))
			By("checking that the vm broker can continue to access the SI")
			serviceInstance.Delete()
		})
	})
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-storage"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleasedEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-google-storage-bucket",
				"default",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.Storage))
			appTwo := apps.Push(apps.WithApp(apps.Storage))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.BindWithParams(appOne, `{"role":"storage.objectAdmin"}`)
			bindingTwo := serviceInstance.BindWithParams(appTwo, `{"role":"storage.objectAdmin"}`)
			apps.Start(appOne, appTwo)

			By("uploading a blob using the first app")
			blobNameOne := random.Hexadecimal()
			blobDataOne := random.Hexadecimal()
			appOne.PUT(blobDataOne, blobNameOne)

			By("downloading the blob using the second app")
			got := appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.BindWithParams(appOne, `{"role":"storage.objectAdmin"}`)
			serviceInstance.BindWithParams(appTwo, `{"role":"storage.objectAdmin"}`)
			apps.Restage(appOne, appTwo)

			By("triggering a no-op update to reapply the terraform for service instance")
			serviceInstance.Update(services.WithParameters(`{}`))

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("deleting bindings created before the update")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.BindWithParams(appOne, `{"role":"storage.objectAdmin"}`)
			serviceInstance.BindWithParams(appTwo, `{"role":"storage.objectAdmin"}`)
			apps.Restage(appOne, appTwo)

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("checking that data can still be written and read")
			blobNameTwo := random.Hexadecimal()
			blobDataTwo := random.Hexadecimal()
			appOne.PUT(blobDataTwo, blobNameTwo)
			got = appTwo.GET(blobNameTwo).String()
			Expect(got).To(Equal(blobDataTwo))

			appOne.DELETE(blobNameOne)
			appOne.DELETE(blobNameTwo)
		})
	})
})
