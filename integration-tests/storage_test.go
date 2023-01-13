package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Storage Bucket", Label("storage"), func() {
	const serviceName = "csb-google-storage-bucket"

	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, serviceName)
		Expect(service.ID).NotTo(BeEmpty())
		Expect(service.Name).NotTo(BeEmpty())
		Expect(service.Tags).To(ConsistOf("gcp", "storage", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeEmpty())
		Expect(service.Metadata.DisplayName).NotTo(BeEmpty())
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("private")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("public-read")}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check region constraints", func() {
			_, err := broker.Provision(serviceName, "public-read", map[string]any{"region": "-Asia-northeast1"})
			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(serviceName, "private", map[string]any{})
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
					HaveKeyWithValue("uniform_bucket_level_access", BeTrue()),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(serviceName, "private", map[string]any{
				"name":                                 "bucket-name",
				"storage_class":                        "STANDARD",
				"region":                               "us",
				"placement_dual_region_data_locations": []string{"us-west1", "us-west2"},
				"public_access_prevention":             "inherited",
				"versioning":                           true,
				"uniform_bucket_level_access":          false,
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
					HaveKeyWithValue("uniform_bucket_level_access", BeFalse()),
				),
			)
		})

		Describe("updating instance", func() {
			var instanceID string

			BeforeEach(func() {
				var err error
				instanceID, err = broker.Provision(serviceName, "public-read", nil)

				Expect(err).NotTo(HaveOccurred())
			})

			DescribeTable(
				"preventing updates with `prohibit_update` as it can force resource replacement or re-creation",
				func(prop string, value any) {
					err := broker.Update(instanceID, serviceName, "public-read", map[string]any{prop: value})

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
			)
		})
	})
})
