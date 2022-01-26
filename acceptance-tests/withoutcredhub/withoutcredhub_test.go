package withoutcredhub_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", func() {
	It("can be accessed by an app", func() {
		env := apps.EnvVar{Name: "CH_CRED_HUB_URL", Value: ""}
		broker := brokers.Create(
			brokers.WithPrefix("csb-no-credhub"),
			brokers.WithEnv(env),
		)
		defer broker.Delete()

		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-google-storage-bucket",
			"private",
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
