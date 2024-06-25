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
	const (
		googleBucketResource    = "google_storage_bucket"
		googleBucketACLResource = "google_storage_bucket_acl"
	)

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
		"versioning":                           true,
		"public_access_prevention":             "fake-public-access-prevention-value",
		"uniform_bucket_level_access":          true,
		"default_kms_key_name":                 "projects/project/locations/location/keyRings/key-ring-name/cryptoKeys/key-name",
		"autoclass":                            false,
		"retention_policy_is_locked":           false,
		"retention_policy_retention_period":    3600,
		"predefined_acl":                       "",
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
					"name":                     Equal("bucket-name"),
					"location":                 Equal("us"),
					"storage_class":            Equal("MULTI_REGIONAL"),
					"labels":                   MatchAllKeys(Keys{"label1": Equal("value1")}),
					"custom_placement_config":  BeEmpty(), // TF internals: It is a []any{} which means no custom_placement_config
					"public_access_prevention": Equal("fake-public-access-prevention-value"),
					"versioning": ConsistOf(
						MatchAllKeys(Keys{
							"enabled": BeTrue(),
						}),
					),
					"uniform_bucket_level_access": BeTrue(),
					"encryption": ConsistOf(
						MatchAllKeys(Keys{
							"default_kms_key_name": Equal("projects/project/locations/location/keyRings/key-ring-name/cryptoKeys/key-name"),
						}),
					),
					"autoclass": BeEmpty(),
					"retention_policy": ConsistOf(
						MatchAllKeys(Keys{
							"is_locked":        BeFalse(),
							"retention_period": BeNumerically("==", 3600),
						}),
					),
				}),
			)
		})

		It("does not create an ACL", func() {
			Expect(AfterValuesForType(plan, googleBucketACLResource)).To(BeNil())
		})
	})

	Context("dual region configuration", func() {
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
					"location":      Equal("us"),
					"storage_class": Equal("STANDARD"),
					"labels":        MatchAllKeys(Keys{"label1": Equal("value1")}),
					"custom_placement_config": ConsistOf(
						MatchAllKeys(Keys{
							"data_locations": ConsistOf("US-WEST1", "US-WEST2"),
						}),
					),
					"versioning": ConsistOf(
						MatchAllKeys(Keys{
							"enabled": BeTrue(),
						}),
					),
				}),
			)
		})
	})

	Context("retention policy", func() {
		When("`retention_policy_is_locked` is nil", func() {
			It("should map to false", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"retention_policy_is_locked":        nil,
					"retention_policy_retention_period": 3600,
				}))
				Expect(AfterValuesForType(plan, googleBucketResource)).To(
					MatchKeys(IgnoreExtras, Keys{
						"retention_policy": ConsistOf(
							MatchAllKeys(Keys{
								"is_locked":        BeFalse(),
								"retention_period": BeNumerically("==", 3600),
							}),
						),
					}),
				)
			})
		})

		When("no retention period `retention_policy`", func() {
			It("should not be defined", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"retention_policy_retention_period": 0,
				}))
				Expect(AfterValuesForType(plan, googleBucketResource)).To(
					MatchKeys(IgnoreExtras, Keys{
						"retention_policy": BeEmpty(), // TF internals: It is a []any{} which means no retention_policy
					}),
				)
			})
		})
	})

	Context("predefined_acl", func() {
		It("creates an ACL", func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"predefined_acl": "publicRead",
			}))

			Expect(AfterValuesForType(plan, googleBucketACLResource)).To(MatchKeys(IgnoreExtras, Keys{
				"bucket":         Equal("bucket-name"),
				"predefined_acl": Equal("publicRead"),
			}))
		})
	})
})
