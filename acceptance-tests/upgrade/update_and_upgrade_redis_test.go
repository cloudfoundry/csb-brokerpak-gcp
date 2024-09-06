package upgrade_test

import (
	"fmt"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeRedisTest", Label("redis"), func() {
	// Context("When upgrading broker version", func() {
	//	It("should continue to work", func() {
	//		By("pushing latest released broker version")
	//		serviceBroker := brokers.Create(
	//			brokers.WithPrefix("csb-upgrade"),
	//			brokers.WithSourceDir(releasedBuildDir),
	//			brokers.WithReleaseEnv(releasedBuildDir),
	//		)
	//		defer serviceBroker.Delete()

	//		By("creating a service")
	//		serviceInstance := services.CreateInstance(
	//			"csb-aws-s3-bucket",
	//			services.WithPlan("default"),
	//			services.WithBroker(serviceBroker),
	//		)
	//		defer serviceInstance.Delete()

	//		By("pushing the unstarted app twice")
	//		appOne := apps.Push(apps.WithApp(apps.S3))
	//		appTwo := apps.Push(apps.WithApp(apps.S3))
	//		defer apps.Delete(appOne, appTwo)

	//		By("binding the apps to the s3 service instance")
	//		bindingOne := serviceInstance.Bind(appOne)
	//		bindingTwo := serviceInstance.Bind(appTwo)
	//		apps.Start(appOne, appTwo)

	//		By("uploading a blob using the first app")
	//		blobNameOne := random.Hexadecimal()
	//		blobDataOne := random.Hexadecimal()
	//		appOne.PUT(blobDataOne, blobNameOne)

	//		By("downloading the blob using the second app")
	//		got := appTwo.GET(blobNameOne).String()
	//		Expect(got).To(Equal(blobDataOne))

	//		//              if os.Getenv("UPGRADE_TO_VM") == "true" {
	//		boshReleasedDir := os.Getenv("BROKER_RELEASE_PATH")
	//		serviceBroker = brokers.CreateVm(
	//			brokers.WithVM(),
	//			brokers.WithName(serviceBroker.Name),
	//			brokers.WithBoshReleaseDir(boshReleasedDir),
	//		)
	//		// broker will register itself in port start, the below is not required.
	//		// serviceBroker.UpdateBrokerToVmBroker()
	//		//              } else {
	//		//                      By("pushing the development version of the broker")
	//		//                      serviceBroker.UpdateBroker(developmentBuildDir)
	//		//              }

	//		By("upgrading the service instance")
	//		serviceInstance.Upgrade()

	//		By("checking that previously written data is accessible")
	//		got = appTwo.GET(blobNameOne).String()
	//		Expect(got).To(Equal(blobDataOne))

	//		By("deleting bindings created before the upgrade")
	//		bindingOne.Unbind()
	//		bindingTwo.Unbind()

	//		By("binding the app to the instance again")
	//		serviceInstance.Bind(appOne)
	//		serviceInstance.Bind(appTwo)
	//		apps.Restage(appOne, appTwo)

	//		By("updating the service instance")
	//		serviceInstance.Update(services.WithParameters(`{}`))
	//	})
	//})
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-redis"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleasedEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")

			serviceInstance := services.CreateInstance(
				"csb-google-redis",
				"basic",
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]any{"instance_id": fmt.Sprintf("test-%s", random.Name(random.WithMaxLength(20)))}),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.Redis))
			appTwo := apps.Push(apps.WithApp(apps.Redis))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)
			apps.Start(appOne, appTwo)

			By("setting a key-value using the first app")
			key1 := random.Hexadecimal()
			value1 := random.Hexadecimal()
			appOne.PUT(value1, key1)

			By("getting the value using the second app")
			Expect(appTwo.GET(key1).String()).To(Equal(value1))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the value using the second app")
			Expect(appTwo.GET(key1).String()).To(Equal(value1))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("triggering a no-op update to reapply the terraform for service instance")
			serviceInstance.Update(services.WithParameters(`{}`))

			By("getting the value using the second app")
			Expect(appTwo.GET(key1).String()).To(Equal(value1))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)
			key2 := random.Hexadecimal()
			value2 := random.Hexadecimal()
			appOne.PUT(value2, key2)
			Expect(appTwo.GET(key2).String()).To(Equal(value2))

			By("getting the value using the second app")
			Expect(appTwo.GET(key1).String()).To(Equal(value1))
		})
	})
})
