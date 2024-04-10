package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	dataprocServiceName        = "csb-google-dataproc"
	dataprocServiceID          = "ebb35d15-8c7a-4c4e-8aa8-d8d751a9d8d3"
	dataprocServiceDisplayName = "Google Cloud Dataproc (Beta)"
	dataprocServiceDescription = "Beta - Dataproc is a fully-managed service for running Apache Spark and Apache Hadoop clusters in a simpler, more cost-efficient way."
	dataprocServiceSupportURL  = "https://cloud.google.com/dataproc/docs/support/getting-support"
	dataprocStandardPlanName   = "standard"
	dataprocStandardPlanID     = "ed8c2ad0-edc7-4f36-a332-fd63d81ec276"
	dataprocHAPlanName         = "ha"
	dataprocHAPlanID           = "71cc321b-3ba3-4f0f-b058-90cfc978e743"
)

var _ = Describe("Dataproc", func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("publishes in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, dataprocServiceName)
		Expect(service.ID).To(Equal(dataprocServiceID))
		Expect(service.Description).To(Equal(dataprocServiceDescription))
		Expect(service.Tags).To(ConsistOf("gcp", "dataproc", "beta"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(dataprocServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(cloudServiceBrokerDocumentationURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(providerDisplayName))
		Expect(service.Metadata.SupportUrl).To(Equal(dataprocServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(dataprocStandardPlanID),
					Name: Equal(dataprocStandardPlanName),
				}),
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(dataprocHAPlanID),
					Name: Equal(dataprocHAPlanName),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check region constraints", func() {
			_, err := broker.Provision(dataprocServiceName, "standard", map[string]any{"region": "-Asia-northeast1"})
			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(dataprocServiceName, "standard", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should allow updating region because it is not flagged as `prohibit_update` and not specified in the plan", func() {
			err := broker.Update(instanceID, dataprocServiceName, "standard", map[string]any{"region": "asia-southeast1"})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
