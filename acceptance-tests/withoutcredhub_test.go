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
		serviceInstance := services.CreateInstance(
			"csb-google-storage-bucket",
			"default",
			services.WithBroker(broker),
		)

		defer serviceInstance.Delete()

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
		got := app.GET(blobName)
		Expect(got).To(Equal(blobData))

		app.DELETE(blobName)
	})
})
