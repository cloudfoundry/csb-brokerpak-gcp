package dataproc_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataproc", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-google-dataproc", "standard")
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		appOne := apps.Push(apps.WithApp(apps.Dataproc))
		defer apps.Delete(appOne)

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(appOne)

		By("starting the apps")
		apps.Start(appOne)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("running a job")
		jobName := random.Hexadecimal()
		appOne.PUT("", jobName)

		By("getting the job status")
		status := appOne.GET(jobName)
		Expect(status).To(Equal("DONE"))

		By("deleting the job")
		appOne.DELETE(jobName)
	})
})
