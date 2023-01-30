package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	storageServiceName             = "csb-google-storage-bucket"
	storageServiceID               = "b247fcde-8a63-11ea-b945-cb26f061f70f"
	storageServiceDisplayName      = "Google Cloud Storage (Beta)"
	storageServiceDocumentationURL = "https://cloud.google.com/storage/docs/overview"
	storageServiceDescription      = "Beta - Google Cloud Storage that uses the Terraform back-end and grants service accounts IAM permissions directly on the bucket."
	storageServiceSupportURL       = "https://cloud.google.com/support/"
	storagePrivatePlanName         = "private"
	storagePrivatePlanID           = "bbc4853e-8a63-11ea-a54e-670ca63cee0b"
	storagePublicPlanName          = "public-read"
	storagePublicPlanID            = "c07f21a6-8a63-11ea-bc1b-d38b123189cb"
)

var _ = Describe("Storage Bucket", Label("storage"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("publishes in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, storageServiceName)
		Expect(service.ID).To(Equal(storageServiceID))
		Expect(service.Description).To(Equal(storageServiceDescription))
		Expect(service.Tags).To(ConsistOf("gcp", "storage", "beta"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(storageServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(storageServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(storageServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(storagePrivatePlanID),
					Name: Equal(storagePrivatePlanName),
				}),
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(storagePublicPlanID),
					Name: Equal(storagePublicPlanName),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check region constraints", func() {
			_, err := broker.Provision(storageServiceName, "public-read", map[string]any{"region": "-Asia-northeast1"})
			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(storageServiceName, "private", map[string]any{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("name", fmt.Sprintf("csb-%s", instanceID)),
					HaveKeyWithValue("storage_class", "MULTI_REGIONAL"),
					HaveKeyWithValue("region", "us"),
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id": Equal(instanceID),
					})),
					HaveKeyWithValue("placement_dual_region_data_locations", BeEmpty()),
					HaveKeyWithValue("public_access_prevention", "enforced"),
					HaveKeyWithValue("versioning", BeFalse()),
					HaveKeyWithValue("uniform_bucket_level_access", BeFalse()),
					HaveKeyWithValue("default_kms_key_name", BeEmpty()),
					HaveKeyWithValue("autoclass", BeFalse()),
					HaveKeyWithValue("retention_policy_is_locked", BeFalse()),
					HaveKeyWithValue("retention_policy_retention_period", BeNumerically("==", 0)),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(storageServiceName, "private", map[string]any{
				"name":                                 "bucket-name",
				"storage_class":                        "STANDARD",
				"region":                               "us",
				"placement_dual_region_data_locations": []string{"us-west1", "us-west2"},
				"public_access_prevention":             "inherited",
				"versioning":                           true,
				"uniform_bucket_level_access":          true,
				"default_kms_key_name":                 "projects/project/locations/location/keyRings/key-ring-name/cryptoKeys/key-name",
				"autoclass":                            true,
				"retention_policy_is_locked":           true,
				"retention_policy_retention_period":    3600,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("name", "bucket-name"),
					HaveKeyWithValue("storage_class", "STANDARD"),
					HaveKeyWithValue("region", "us"),
					HaveKeyWithValue("placement_dual_region_data_locations", ConsistOf("us-west1", "us-west2")),
					HaveKeyWithValue("public_access_prevention", "inherited"),
					HaveKeyWithValue("versioning", BeTrue()),
					HaveKeyWithValue("uniform_bucket_level_access", BeTrue()),
					HaveKeyWithValue("default_kms_key_name", "projects/project/locations/location/keyRings/key-ring-name/cryptoKeys/key-name"),
					HaveKeyWithValue("autoclass", BeTrue()),
					HaveKeyWithValue("retention_policy_is_locked", BeTrue()),
					HaveKeyWithValue("retention_policy_retention_period", BeNumerically("==", 3600)),
				),
			)
		})

		Describe("updating instance", func() {
			var instanceID string

			BeforeEach(func() {
				var err error
				instanceID, err = broker.Provision(storageServiceName, "public-read", nil)

				Expect(err).NotTo(HaveOccurred())
			})

			DescribeTable(
				"preventing updates with `prohibit_update` as it can force resource replacement or re-creation",
				func(prop string, value any) {
					err := broker.Update(instanceID, storageServiceName, "public-read", map[string]any{prop: value})

					Expect(err).To(MatchError(
						ContainSubstring(
							"attempt to update parameter that may result in service instance re-creation and data loss",
						),
					))

					const initialProvisionInvocation = 1
					Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
				},
				Entry("region", "region", "no-matter-what-region"),
				Entry("placement_dual_region_data_locations", "placement_dual_region_data_locations", []string{"us-west1", "us-west2"}),
				Entry("autoclass", "autoclass", true),
			)
		})
	})
})
