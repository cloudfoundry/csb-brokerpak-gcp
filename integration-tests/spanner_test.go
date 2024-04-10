package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	spannerServiceName        = "csb-google-spanner"
	spannerServiceID          = "528667ba-8e4e-11ea-954d-47e3237b9109"
	spannerServiceDisplayName = "Google Cloud Spanner (Beta)"
	spannerServiceDescription = "Beta - Fully managed, scalable, relational database service for regional and global application data."
	spannerServiceSupportURL  = "https://cloud.google.com/support/"
	spannerSmallPlanName      = "small"
	spannerSmallPlanID        = "706659ba-8e4f-11ea-a91e-4328fa08a19b"
	spannerMediumPlanName     = "medium"
	spannerMediumPlanID       = "7564d13a-8e4f-11ea-ac11-5b314921fb4c"
	spannerLargePlanName      = "large"
	spannerLargePlanID        = "7c172a64-8e4f-11ea-a471-731172c1c00f"
)

var _ = Describe("Spanner", func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("publishes in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, spannerServiceName)
		Expect(service.ID).To(Equal(spannerServiceID))
		Expect(service.Description).To(Equal(spannerServiceDescription))
		Expect(service.Tags).To(ConsistOf("gcp", "spanner", "beta"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(spannerServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(cloudServiceBrokerDocumentationURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(providerDisplayName))
		Expect(service.Metadata.SupportUrl).To(Equal(spannerServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(spannerSmallPlanID),
					Name: Equal(spannerSmallPlanName),
				}),
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(spannerMediumPlanID),
					Name: Equal(spannerMediumPlanName),
				}),
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(spannerLargePlanID),
					Name: Equal(spannerLargePlanName),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check region constraints", func() {
			_, err := broker.Provision(spannerServiceName, "small", map[string]any{"region": "-Asia-northeast1"})
			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(spannerServiceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should allow updating region because it is not flagged as `prohibit_update` and not specified in the plan", func() {
			err := broker.Update(instanceID, spannerServiceName, "small", map[string]any{"region": "asia-southeast1"})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
