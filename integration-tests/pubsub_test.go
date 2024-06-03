package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	pubsubServiceName        = "csb-google-pubsub"
	pubsubServiceID          = "59c8535c-d068-4078-b293-a368b09a1a32"
	pubsubServiceDisplayName = "Google Pub/Sub (Beta)"
	pubsubServiceDescription = "Beta - Google Pub/Sub is an asynchronous and scalable messaging service that decouples services producing messages from services processing those messages."
	pubsubServiceSupportURL  = "https://cloud.google.com/support/"
	pubsubDefaultPlanName    = "default"
	pubsubDefaultPlanID      = "0690bcd8-e29e-4317-9387-f8152501403d"
)

var customPubSubPlans = []map[string]any{
	customPubSubPlan,
}

var customPubSubPlan = map[string]any{
	"name": pubsubDefaultPlanName,
	"id":   pubsubDefaultPlanID,
	"metadata": map[string]any{
		"displayName": pubsubServiceDisplayName,
	},
	"labels": map[string]any{
		"label1": "label1",
		"label2": "label2",
	},
}

var _ = Describe("PubSub", Label("pubsub"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("publishes in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, pubsubServiceName)
		Expect(service.ID).To(Equal(pubsubServiceID))
		Expect(service.Description).To(Equal(pubsubServiceDescription))
		Expect(service.Tags).To(ConsistOf("gcp", "pubsub", "google-pubsub", "beta"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(pubsubServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(cloudServiceBrokerDocumentationURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(providerDisplayName))
		Expect(service.Metadata.SupportUrl).To(Equal(pubsubServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(pubsubDefaultPlanID),
					Name: Equal(pubsubDefaultPlanName),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		DescribeTable("should check property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(pubsubServiceName, pubsubDefaultPlanName, params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"ack_deadline minimum is 10 seconds",
				map[string]any{"ack_deadline": 5},
				"ack_deadline: Must be greater than or equal to 10",
			),
			Entry(
				"ack_deadline maximum is 600 seconds",
				map[string]any{"ack_deadline": 700},
				"ack_deadline: Must be less than or equal to 600",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(pubsubServiceName, pubsubDefaultPlanName, map[string]any{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("topic_name", fmt.Sprintf("csb-topic-%s", instanceID)),
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id": Equal(instanceID),
					})),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(pubsubServiceName, pubsubDefaultPlanName, map[string]any{
				"topic_name":        "test-topic-name",
				"subscription_name": "test-subscription-name",
				"ack_deadline":      600,
				"push_endpoint":     "https://example.test/push",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("topic_name", "test-topic-name"),
					HaveKeyWithValue("subscription_name", "test-subscription-name"),
					HaveKeyWithValue("ack_deadline", BeNumerically("==", 600)),
					HaveKeyWithValue("push_endpoint", "https://example.test/push"),
				),
			)
		})
	})
})
