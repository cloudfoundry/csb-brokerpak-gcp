package integration_tests

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var _ = Describe("Mysql", func() {
	AfterEach(func() {
		Expect(mockTerraform.Reset()).NotTo(HaveOccurred())
	})
	It("publish mysql in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())
		service := testframework.FindService(catalog, "csb-google-mysql")
		Expect(service.Plans).To(HaveLen(3))
		Expect(service.Tags).To(ContainElement("preview"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
	})

	It("should provision small plan", func() {
		broker.Provision("csb-google-mysql", "small", nil)

		invocations, err := mockTerraform.ApplyInvocations()
		Expect(err).NotTo(HaveOccurred())
		Expect(invocations).To(HaveLen(1))

		contents, err := invocations[0].TFVarsContents()
		Expect(err).NotTo(HaveOccurred())
		Expect(replaceGUIDs(contents)).To(MatchJSON(getResultContents("mysql-result")))
	})

	It("user should be able to update database name", func() {
		broker.Provision("csb-google-mysql", "small", map[string]interface{}{"db_name": "foobar"})
		Expect(mockTerraform.FirstTerraformInvocationVars()).To(HaveKeyWithValue("db_name", "foobar"))
	})

	It("user should be able to update database name", func() {
		broker.Provision("csb-google-mysql", "small", map[string]interface{}{"db_name": "foobar"})
		Expect(mockTerraform.FirstTerraformInvocationVars()).To(HaveKeyWithValue("db_name", "foobar"))
	})

	It("user should not be allowed to change mysql cores", func() {
		_, err := broker.Provision("csb-google-mysql", "small", map[string]interface{}{"cores": 5})
		Expect(err).To(MatchError(ContainSubstring("plan defined properties cannot be changed: cores")))
	})

	It("should validate region", func() {
		_, err := broker.Provision("csb-google-mysql", "small", map[string]interface{}{"region": "invalid-region"})
		Expect(err).To(MatchError(ContainSubstring("region must be one of the following:")))
	})

	It("should validate instance name length", func() {
		_, err := broker.Provision("csb-google-mysql", "small", map[string]interface{}{"instance_name": "2smol"})
		Expect(err).To(MatchError(ContainSubstring("instance_name: String length must be greater than or equal to 6")))
	})
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
