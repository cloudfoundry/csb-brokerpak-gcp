package upgrade_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
	"csbbrokerpakgcp/acceptance-tests/helpers/gcloud"
	"csbbrokerpakgcp/acceptance-tests/helpers/plans"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
	"fmt"
	"strings"

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

			By("creating a service")
			serviceOffering := "csb-google-postgres"
			servicePlan := "small"
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
			appOne := apps.Push(apps.WithApp(apps.PostgreSQL))
			appTwo := apps.Push(apps.WithApp(apps.PostgreSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			instanceName := fmt.Sprintf("csb-postgres-%s", serviceInstance.GUID())
			ip := strings.TrimSpace(string(gcloud.GCP("sql", "instances", "describe", instanceName, "--format=value(ipAddresses[0].ipAddress)")))
			By("verifying network reachability to the database at " + ip)
			cf.Run("ssh", serviceBroker.Name, "-c", fmt.Sprintf("nc -zvw 10 %s 5432", ip))

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
			appOne.PUTf(valueOne, "%s/%s", schema, keyOne)

			By("getting the value using the second app")
			got := appTwo.GETf("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("checking previously written data still accessible")
			got = appTwo.GETf("%s/%s", schema, keyOne).String()
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
			got = appTwo.GETf("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("deleting bindings created before the update")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking previously written data still accessible")
			got = appTwo.GETf("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUTf(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GETf("%s/%s", schema, keyTwo).String()
			Expect(got).To(Equal(valueTwo))
		})
	})
})
