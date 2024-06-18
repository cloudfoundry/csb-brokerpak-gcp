package terraformtests

import (
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "csbbrokerpakgcp/terraform-tests/helpers"
)

var _ = Describe("PubSub", Label("pubsub-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
		defaultVars           map[string]any
	)

	BeforeAll(func() {
		defaultVars = map[string]any{
			"credentials":                      googleCredentials,
			"project":                          googleProject,
			"labels":                           map[string]string{"label1": "value1"},
			"topic_name":                       "test-topic-name",
			"subscription_name":                "",
			"ack_deadline":                     10,
			"push_endpoint":                    "",
			"topic_message_retention_duration": "",
			"topic_kms_key_name":               "",
			"subscription_message_retention_duration":   "",
			"subscription_retain_acked_messages":        false,
			"subscription_expiration_policy":            nil,
			"subscription_retry_policy_minimum_backoff": "",
			"subscription_retry_policy_maximum_backoff": "",
			"subscription_enable_message_ordering":      false,
			"subscription_enable_exactly_once_delivery": false,
		}
		terraformProvisionDir = path.Join(workingDir, "pubsub/provision")
		Init(terraformProvisionDir)
	})

	Context("with default values", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(1))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"google_pubsub_topic",
			))
		})

		It("should create a topic with the right values", func() {
			Expect(AfterValuesForType(plan, "google_pubsub_topic")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name": Equal("test-topic-name"),
				}),
			)
		})
	})

	When("subscription name is passed", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"subscription_name": "test-subscription-name",
			}))
		})

		It("should create a subscription with the right values", func() {
			Expect(plan.ResourceChanges).To(HaveLen(2))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"google_pubsub_topic",
				"google_pubsub_subscription",
			))
			Expect(AfterValuesForType(plan, "google_pubsub_subscription")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                 Equal("test-subscription-name"),
					"topic":                Equal("test-topic-name"),
					"ack_deadline_seconds": BeNumerically("==", 10),
				}),
			)
		})
	})

	When("a push subscription is requested", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"subscription_name": "test-subscription-name",
				"push_endpoint":     "https://example.com/push",
			}))
		})

		It("should create a subscription with the right values", func() {
			Expect(plan.ResourceChanges).To(HaveLen(2))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"google_pubsub_topic",
				"google_pubsub_subscription",
			))
			Expect(AfterValuesForType(plan, "google_pubsub_subscription")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                 Equal("test-subscription-name"),
					"topic":                Equal("test-topic-name"),
					"ack_deadline_seconds": BeNumerically("==", 10),
					"push_config": ContainElement(MatchKeys(IgnoreExtras, Keys{
						"push_endpoint": Equal("https://example.com/push"),
					})),
				}),
			)
		})
	})

	When("additional topic and subscription properties are set", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"topic_message_retention_duration":          "604800s",
				"topic_kms_key_name":                        "projects/my-project/locations/us/keyRings/my-keyring/cryptoKeys/my-key",
				"subscription_name":                         "test-subscription-name",
				"subscription_message_retention_duration":   "604800s",
				"subscription_retain_acked_messages":        true,
				"subscription_expiration_policy":            "604800s",
				"subscription_retry_policy_minimum_backoff": "10s",
				"subscription_retry_policy_maximum_backoff": "600s",
				"subscription_enable_message_ordering":      true,
				"subscription_enable_exactly_once_delivery": true,
			}))
		})

		It("should create a topic and subscription with the right values", func() {
			Expect(plan.ResourceChanges).To(HaveLen(2))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"google_pubsub_topic",
				"google_pubsub_subscription",
			))

			Expect(AfterValuesForType(plan, "google_pubsub_topic")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                       Equal("test-topic-name"),
					"message_retention_duration": Equal("604800s"),
					"kms_key_name":               Equal("projects/my-project/locations/us/keyRings/my-keyring/cryptoKeys/my-key"),
				}),
			)

			Expect(AfterValuesForType(plan, "google_pubsub_subscription")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                       Equal("test-subscription-name"),
					"topic":                      Equal("test-topic-name"),
					"ack_deadline_seconds":       BeNumerically("==", 10),
					"message_retention_duration": Equal("604800s"),
					"retain_acked_messages":      BeTrue(),
					"expiration_policy": ConsistOf(
						MatchAllKeys(Keys{
							"ttl": Equal("604800s"),
						}),
					),
					"retry_policy": ConsistOf(
						MatchAllKeys(Keys{
							"minimum_backoff": Equal("10s"),
							"maximum_backoff": Equal("600s"),
						}),
					),
					"enable_message_ordering":      BeTrue(),
					"enable_exactly_once_delivery": BeTrue(),
				}),
			)
		})
	})
})
