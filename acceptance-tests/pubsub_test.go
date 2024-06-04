package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/matchers"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
	"csbbrokerpakgcp/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PubSub", Label("pubsub"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-google-pubsub",
			"default",
			services.WithParameters(map[string]any{"subscription_name": random.Name()}))
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		publisherApp := apps.Push(apps.WithApp(apps.PubSubApp))
		subscriberApp := apps.Push(apps.WithApp(apps.PubSubApp))
		defer apps.Delete(publisherApp, subscriberApp)

		By("binding the apps to the storage service instance")
		binding := serviceInstance.BindWithParams(publisherApp, `{"role":"pubsub.editor"}`)
		serviceInstance.BindWithParams(subscriberApp, `{"role":"pubsub.editor"}`)

		By("starting the apps")
		apps.Start(publisherApp, subscriberApp)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("publishing a message with the publisher app")
		messageData := random.Hexadecimal()
		publisherApp.PUT(messageData, "")

		By("retrieving a message with the subscriber app")
		got := subscriberApp.GET("").String()
		Expect(got).To(Equal(messageData))
	})
})
