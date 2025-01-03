package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", Label("withoutcredhub"), func() {
	It("can be accessed by an app", func() {
		broker := brokers.Create(
			brokers.WithPrefix("csb-no-credhub"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: "CH_CRED_HUB_URL", Value: ""}),
		)
		defer broker.Delete()

		By("creating a service instance")
		serviceOffering := "csb-google-storage-bucket"
		servicePlan := "default"
		serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
		defer services.Delete(serviceName)
		// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
		// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
		// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
		// (which can only have 5 instances at any time) to prevent subsequent test failures.
		serviceInstance := services.CreateInstance(
			serviceOffering,
			servicePlan,
			services.WithBroker(broker),
			services.WithName(serviceName),
		)

		By("pushing the unstarted app")
		app := apps.Push(apps.WithApp(apps.Storage))
		defer apps.Delete(app)

		By("binding the app to the storage service instance")
		binding := serviceInstance.BindWithParams(app, `{"role":"storage.objectAdmin"}`)

		By("starting the app")
		apps.Start(app)

		By("checking that the app environment does not a credhub reference for credentials")
		Expect(binding.Credential()).NotTo(matchers.HaveCredHubRef)

		By("uploading a blob")
		blobName := random.Hexadecimal()
		blobData := random.Hexadecimal()
		app.PUT(blobData, blobName)

		By("downloading the blob")
		got := app.GET(blobName).String()
		Expect(got).To(Equal(blobData))

		app.DELETE(blobName)
	})
})
