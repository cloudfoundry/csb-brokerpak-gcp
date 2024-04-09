package upgrade_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("UpgradeSpannerTest", Label("spanner"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-spanner"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleasedEnv(releasedBuildDir),
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
			Expect(appOne.GET(key).String()).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			// Story: https://www.pivotaltracker.com/story/show/187392169
			// After upgrading the Google Terraform provider from v4 to v5
			// (https://github.com/cloudfoundry/csb-brokerpak-gcp/pull/1126
			// https://github.com/cloudfoundry/csb-brokerpak-gcp/commit/692ef080fed595a211fc0952212a230acc7657be)
			// it was observed that the upgrade of a spanner instance failed with error:
			//
			//   message:   upgrade failed: Error: Error updating Instance "cloud-service-broker/csb-spanner-235ac12a-ddaf-4305-88d1-8cfe0a837130": googleapi: Error 400: Invalid UpdateInstance request. Details: [ { "@type":
			//              "type.googleapis.com/google.rpc.BadRequest", "fieldViolations": [ { "description": "Must specify a non-empty field mask.", "field": "field_mask" } ] } ] with google_spanner_instance.spanner_instance, on main.tf
			//              line 4, in resource "google_spanner_instance" "spanner_instance": 4: resource "google_spanner_instance" "spanner_instance" { exit status 1
			//
			// But a second attempt to upgrade would succeed. This is likely because the underlying first "terraform apply"
			// synchronises the state, and a second "terraform apply" then notes that there's nothing to update. Spanner
			// is an unsupported (Beta) service, we are not aware of any users, and there's a simple work-around of
			// re-running the upgrade, so we chose not to investigate any further.
			if serviceInstance.UpgradeAvailable() {
				By("attempting to upgrade the service instance a first time")

				session := cf.Start("upgrade-service", serviceInstance.Name, "--force", "--wait")
				Eventually(session, time.Hour).Should(gexec.Exit(), func() string {
					out, _ := cf.Run("service", serviceInstance.Name)
					return out
				})

				if session.ExitCode() != 0 {
					By("attempting to upgrade the service instance a second time")
					serviceInstance.Upgrade()
				} else {
					By("noting that the upgrade succeeded")
				}
			} else {
				By("noting that no upgrade is available")
			}

			By("checking previously written data still accessible")
			Expect(appOne.GET(key).String()).To(Equal(value))
		})
	})
})
