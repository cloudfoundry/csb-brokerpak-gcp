package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var customMySQLPlan = map[string]any{
	"name":                  "custom-plan",
	"id":                    "9daa07f1-78e8-4bda-9efe-91576102c30d",
	"description":           "custom plan defined by customer",
	"mysql_version":         "MYSQL_5_7",
	"credentials":           "plan_cred",
	"project":               "plan_project",
	"authorized_network":    "plan_authorized_network",
	"authorized_network_id": "plan_authorized_network_id",
	"require_ssl":           false,
	"metadata": map[string]any{
		"displayName": "custom plan defined by customer (beta)",
	},
}

var _ = Describe("Mysql", func() {
	AfterEach(func() {
		Expect(mockTerraform.Reset()).NotTo(HaveOccurred())
	})

	It("should publish mysql in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, "csb-google-mysql")
		Expect(service.Plans).To(HaveLen(4))
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf([]string{"gcp", "mysql", "beta"}))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(marshall(service.Plans)).To(MatchJSON(getResultContents("mysql-plans")))
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

	DescribeTable("property constraints",
		func(params map[string]any, expectedErrorMsg string) {
			_, err := broker.Provision("csb-google-mysql", customMySQLPlan["name"].(string), params)

			Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
		},
		Entry("cores maximum value is 64", map[string]any{"cores": 65}, "cores: Must be a multiple of 2; cores: Must be less than or equal to 64"),
		Entry("cores minimum value is 1", map[string]any{"cores": 0}, "cores: Must be greater than or equal to 1"),
		Entry("cores multiple of 2", map[string]any{"cores": 3}, "cores: Must be a multiple of 2"),
		Entry("storage capacity maximum value is 4096", map[string]any{"storage_gb": 4097}, "storage_gb: Must be less than or equal to 4096"),
		Entry("storage capacity minimum value is 10", map[string]any{"storage_gb": 9}, "storage_gb: Must be greater than or equal to 10"),
		Entry("instance name maximum length is 98 characters", map[string]any{"instance_name": stringOfLen(99)}, "instance_name: String length must be less than or equal to 98"),
		Entry("instance name minimum length is 6 characters", map[string]any{"instance_name": stringOfLen(5)}, "instance_name: String length must be greater than or equal to 6"),
		Entry("instance name invalid characters", map[string]any{"instance_name": ".aaaaa"}, "instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'"),
		Entry("database name maximum length is 64 characters", map[string]any{"db_name": stringOfLen(65)}, "db_name: String length must be less than or equal to 64"),
		Entry("invalid region", map[string]any{"region": "invalid-region"}, "region must be one of the following:"),
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

func stringOfLen(length int) string {
	return strings.Repeat("a", length)
}

func marshall(element any) []byte {
	b, err := json.Marshal(element)
	Expect(err).NotTo(HaveOccurred())
	return b
}
