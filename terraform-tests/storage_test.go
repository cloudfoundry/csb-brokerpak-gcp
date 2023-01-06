package terraformtests

import (
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "csbbrokerpakgcp/terraform-tests/helpers"
)

var _ = Describe("storage", Label("storage-terraform"), Ordered, func() {
	const googleBucketResource = "google_storage_bucket"

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"credentials":                          googleCredentials,
		"project":                              googleProject,
		"region":                               "us",
		"labels":                               map[string]string{"label1": "value1"},
		"name":                                 "bucket-name",
		"storage_class":                        "MULTI_REGIONAL",
		"placement_dual_region_data_locations": []string{},
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "storage/provision")
		Init(terraformProvisionDir)
	})

	Context("default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("maps parameters to corresponding values", func() {
			Expect(AfterValuesForType(plan, googleBucketResource)).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                    Equal("bucket-name"),
					"location":                Equal("US"),
					"storage_class":           Equal("MULTI_REGIONAL"),
					"labels":                  MatchKeys(0, Keys{"label1": Equal("value1")}),
					"custom_placement_config": BeEmpty(), // TF internals: It is a []any{} which means no custom_placement_config
				}),
			)
		})
	})

	FContext("dual region configuration", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"storage_class":                        "STANDARD",
				"placement_dual_region_data_locations": []string{"us-west1", "us-west2"},
			}))
		})

		It("maps parameters to corresponding values", func() {
			Expect(AfterValuesForType(plan, googleBucketResource)).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":          Equal("bucket-name"),
					"location":      Equal("US"),
					"storage_class": Equal("STANDARD"),
					"labels":        MatchKeys(0, Keys{"label1": Equal("value1")}),
					"custom_placement_config": ConsistOf(
						MatchKeys(0, Keys{
							"data_locations": ConsistOf("us-west1", "us-west2"),
						}),
					),
				}),
			)
		})
	})
})
