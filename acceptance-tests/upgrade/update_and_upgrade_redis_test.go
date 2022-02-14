package upgrade_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeRedisTest", Label("redis"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-redis"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithEnv(apps.EnvVar{
					Name:  "GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS",
					Value: "",
				}),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-google-redis",
				"basic",
				services.WithBroker(serviceBroker),
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
			got := appTwo.GET(key1)
			Expect(got).To(Equal(value1))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(
				developmentBuildDir,
				apps.EnvVar{
					Name:  "GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS",
					Value: `[{"name":"small","id":"85b27a04-8695-11ea-818a-274131861b81","description":"PostgreSQL v11, shared CPU, minumum 0.6GB ram, 10GB storage","display_name":"small","cores":0.6,"postgres_version":"POSTGRES_11","storage_gb":10},{"name":"medium","id":"b41ee300-8695-11ea-87df-cfcb8aecf3bc","description":"PostgreSQL v11, shared CPU, minumum 1.7GB ram, 20GB storage","display_name":"medium","cores":1.7,"postgres_version":"POSTGRES_11","storage_gb":20},{"name":"large","id":"2a57527e-b025-11ea-b643-bf3bcf6d055a","description":"PostgreSQL v11, minumum 8 cores, minumum 8GB ram, 50GB storage","display_name":"large","cores":8,"postgres_version":"POSTGRES_11","storage_gb":50}]`,
				})

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
			Expect(appTwo.GET(key2)).To(Equal(value2))

			By("getting the value using the second app")
			Expect(appTwo.GET(key1)).To(Equal(value1))

		})
	})
})
