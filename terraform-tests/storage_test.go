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
		"credentials":   googleCredentials,
		"project":       googleProject,
		"region":        "us-central1",
		"labels":        map[string]string{"label1": "value1"},
		"name":          "bucket-name",
		"storage_class": "MULTI_REGIONAL",
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
					"name":          Equal("bucket-name"),
					"location":      Equal("US-CENTRAL1"),
					"storage_class": Equal("MULTI_REGIONAL"),
					"labels":        MatchKeys(0, Keys{"label1": Equal("value1")}),
				}),
			)
		})
	})
})
