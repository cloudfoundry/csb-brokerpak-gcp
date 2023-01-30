package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	bigqueryServiceName             = "csb-google-bigquery"
	bigqueryServiceID               = "3d4b5b0c-931d-11ea-a02b-cb6a223f4ab2"
	bigqueryServiceDisplayName      = "Google Big Query (Beta)"
	bigqueryServiceDocumentationURL = "https://cloud.google.com/bigquery/docs/"
	bigqueryServiceDescription      = "Beta - A fast, economical and fully managed data warehouse for large-scale data analytics."
	bigqueryServiceSupportURL       = "https://cloud.google.com/support/"
	bigqueryStandardPlanName        = "standard"
	bigqueryStandardPlanID          = "481212b0-931d-11ea-b054-535fa8f91417"
)

var _ = Describe("BigQuery", func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("publishes in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, bigqueryServiceName)
		Expect(service.ID).To(Equal(bigqueryServiceID))
		Expect(service.Description).To(Equal(bigqueryServiceDescription))
		Expect(service.Tags).To(ConsistOf("gcp", "bigquery", "beta"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(bigqueryServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(bigqueryServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(bigqueryServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(bigqueryStandardPlanID),
					Name: Equal(bigqueryStandardPlanName),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check region constraints", func() {
			_, err := broker.Provision(bigqueryServiceName, "standard", map[string]any{"region": "-Asia-northeast1"})
			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(bigqueryServiceName, "standard", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, bigqueryServiceName, "standard", map[string]any{"region": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
		})
	})
})
