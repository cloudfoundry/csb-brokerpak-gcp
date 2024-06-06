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
		Expect(got).To(Equal(messageData), "Received message matched published message")
	})

	When("using the legacy broker", func() {
		It("can continue using the same app with CSB service instance", func() {
			By("creating a legacy service instance")
			legacySubscription := random.Name()
			legacyInstance := services.CreateInstance(
				"google-pubsub",
				"default",
				services.WithBrokerName("legacy-gcp-broker"),
				services.WithParameters(map[string]any{"subscription_name": legacySubscription}))
			defer legacyInstance.Delete()

			By("pushing the unstarted app twice")
			publisherApp := apps.Push(apps.WithApp(apps.PubSubApp))
			subscriberApp := apps.Push(apps.WithApp(apps.PubSubApp))
			defer apps.Delete(publisherApp, subscriberApp)

			By("binding the apps to the storage service instance")
			legacyPubBinding := legacyInstance.BindWithParams(publisherApp, `{"role":"pubsub.editor"}`)
			legacySubBinding := legacyInstance.BindWithParams(subscriberApp, `{"role":"pubsub.editor"}`)

			By("starting the apps")
			apps.Start(publisherApp, subscriberApp)

			By("publishing a message with the publisher app")
			messageData := random.Hexadecimal()
			publisherApp.PUT(messageData, "")

			By("retrieving a message with the subscriber app")
			got := subscriberApp.GET("").String()
			Expect(got).To(Equal(messageData), "Received message matched published message")

			By("creating a CSB service instance")
			CSBServiceInstance := services.CreateInstance(
				"csb-google-pubsub",
				"default",
				services.WithParameters(map[string]any{"subscription_name": random.Name()}))
			defer CSBServiceInstance.Delete()

			By("unbinding the apps from legacy service instance and binding them to CSB instance")
			legacyPubBinding.Unbind()
			legacySubBinding.Unbind()
			CSBServiceInstance.BindWithParams(publisherApp, `{"role":"pubsub.editor"}`)
			CSBServiceInstance.BindWithParams(subscriberApp, `{"role":"pubsub.editor"}`)

			By("starting the apps")
			apps.Restage(publisherApp, subscriberApp)

			By("publishing a message with the publisher app")
			newMessageData := random.Hexadecimal()
			publisherApp.PUT(newMessageData, "")

			By("retrieving a message with the subscriber app")
			result := subscriberApp.GET("").String()
			Expect(result).To(Equal(newMessageData), "Received message matched published message")
		})
	})
})
