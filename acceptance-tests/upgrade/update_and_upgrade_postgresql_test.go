package upgrade_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/plans"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradePostgreSQLTest", Label("postgresql"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-postgresql"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleasedEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			// We have to start the defered delete *before* creating the service.
			// If there is an error in a creating the service then services.CreateInstance won't return
			// and we may have a failed creation in the database attached to the broker we just created,
			// preventing deleting the broker. Calling `cf delete-service` still needs to be done.
			serviceName := "csb-google-postgres"
			defer services.Delete(serviceName)
			By("creating a service")
			serviceInstance := services.CreateInstance(
				serviceName,
				"small",
				services.WithBroker(serviceBroker),
			)

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.PostgreSQL))
			appTwo := apps.Push(apps.WithApp(apps.PostgreSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("creating a schema using the first app")
			schema := random.Name(random.WithMaxLength(10))
			appOne.PUT("", schema)

			By("setting a key-value using the first app")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			appOne.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value using the second app")
			got := appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable("small", serviceName, serviceBroker.Name))

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("updating the instance plan")
			serviceInstance.Update(services.WithParameters(`{}`))

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("deleting bindings created before the update")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GET("%s/%s", schema, keyTwo).String()
			Expect(got).To(Equal(valueTwo))
		})
	})
})
