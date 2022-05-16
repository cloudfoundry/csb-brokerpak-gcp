package integration_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var customMySQLPlan = map[string]interface{}{
	"name":                  "custom-plan",
	"id":                    "9daa07f1-78e8-4bda-9efe-91576102c30d",
	"description":           "custom plan defined by customer",
	"display_name":          "custom plan defined by customer (beta)",
	"mysql_version":         "MYSQL_5_7",
	"credentials":           "plan_cred",
	"project":               "plan_project",
	"authorized_network":    "plan_authorized_network",
	"authorized_network_id": "plan_authorized_network_id",
	"require_ssl":           false,
}

var _ = Describe("Mysql", func() {
	AfterEach(func() {
		Expect(mockTerraform.Reset()).NotTo(HaveOccurred())
	})

	It("should publish mysql in the catalog", func() {
		expectedPlans := []string{"small", "medium", "large", customMySQLPlan["name"].(string)}

		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, "csb-google-mysql")
		Expect(service.Plans).To(HaveLen(4))
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf([]string{"gcp", "mysql", "beta"}))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		for _, plan := range service.Plans {
			Expect(plan.Name).To(BeElementOf(expectedPlans))
		}
	})

	Describe("provisioning", func() {
		It("should provision small plan", func() {
			broker.Provision("csb-google-mysql", "small", nil)

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))

			contents, err := invocations[0].TFVarsContents()
			Expect(err).NotTo(HaveOccurred())
			Expect(replaceGUIDs(contents)).To(MatchJSON(getResultContents("mysql-result")))
		})

		It("should allow setting of database name", func() {
			broker.Provision("csb-google-mysql", "small", map[string]interface{}{"db_name": "foobar"})

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(HaveKeyWithValue("db_name", "foobar"))
		})

		It("should not allow changing of plan defined properties", func() {
			_, err := broker.Provision("csb-google-mysql", "small", map[string]interface{}{"cores": 5})

			Expect(err).To(MatchError(ContainSubstring("plan defined properties cannot be changed: cores")))
		})
	})

	type constraint struct {
		params           map[string]interface{}
		expectedErrorMsg string
	}

	DescribeTable("property constraints",
		func(p constraint) {
			_, err := broker.Provision("csb-google-mysql", customMySQLPlan["name"].(string), p.params)

			Expect(err).To(MatchError(ContainSubstring(p.expectedErrorMsg)))
		},
		Entry("should not allow setting the number of cores because the maximum value is 64", constraint{
			params:           map[string]interface{}{"cores": 65},
			expectedErrorMsg: "cores: Must be a multiple of 2; cores: Must be less than or equal to 64",
		}),
		Entry("should not allow setting the number of cores because the minimum value is 1", constraint{
			params:           map[string]interface{}{"cores": 0},
			expectedErrorMsg: "cores: Must be greater than or equal to 1",
		}),
		Entry("should not allow setting the number of cores because it is not a multiple of 2", constraint{
			params:           map[string]interface{}{"cores": 3},
			expectedErrorMsg: "cores: Must be a multiple of 2",
		}),
		Entry("should not allow setting the storage capacity because the maximum value is 4096", constraint{
			params:           map[string]interface{}{"storage_gb": 4097},
			expectedErrorMsg: "storage_gb: Must be less than or equal to 4096",
		}),
		Entry("should not allow setting the storage capacity because the minimum value is 10", constraint{
			params:           map[string]interface{}{"storage_gb": 9},
			expectedErrorMsg: "storage_gb: Must be greater than or equal to 10",
		}),
		Entry("should not allow setting the instance name because the maximum length is 98 characters", constraint{
			params:           map[string]interface{}{"instance_name": generateString(99)},
			expectedErrorMsg: "instance_name: String length must be less than or equal to 98",
		}),
		Entry("should not allow setting the instance name because the minimum length is 6 characters", constraint{
			params:           map[string]interface{}{"instance_name": generateString(5)},
			expectedErrorMsg: "instance_name: String length must be greater than or equal to 6",
		}),
		Entry("should not allow setting the instance name because of invalid characters", constraint{
			params:           map[string]interface{}{"instance_name": ".aaaaa"},
			expectedErrorMsg: "instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'",
		}),
		Entry("should not allow setting the name of the database because the maximum length is 64 characters", constraint{
			params:           map[string]interface{}{"db_name": generateString(65)},
			expectedErrorMsg: "db_name: String length must be less than or equal to 64",
		}),
		Entry("should not allow setting an invalid region", constraint{
			params:           map[string]interface{}{"region": "invalid-region"},
			expectedErrorMsg: "region must be one of the following:",
		}),
	)
})

var guidRegex = regexp.MustCompile("[{]?[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[}]?")

func replaceGUIDs(contents string) string {
	return guidRegex.ReplaceAllString(contents, "GUID")
}

func getResultContents(name string) string {
	contents, err := os.ReadFile(getResultFilePath(name))
	Expect(err).NotTo(HaveOccurred())
	return string(contents)
}

func getResultFilePath(name string) string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Join(filepath.Dir(file), "results", name+".json")
}

func generateString(length int) string {
	return strings.Repeat("a", length)
}
