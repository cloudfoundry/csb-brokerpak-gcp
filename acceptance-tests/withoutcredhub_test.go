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
		// We have to create the defered delete *before* creating the service.
		// If there is an error in a creating the service then services.CreateInstance won't return
		// and we may have a failed creation in the database attached to the broker we just created,
		// preventing deleting the broker. Calling `cf delete-service` still needs to be done.
		serviceName := "csb-google-storage-bucket"
		defer services.Delete(serviceName)
		serviceInstance := services.CreateInstance(
			serviceName,
			"default",
			services.WithBroker(broker),
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
