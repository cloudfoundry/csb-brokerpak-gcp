package terraformtests

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cp "github.com/otiai10/copy"
	"golang.org/x/exp/maps"
)

func TestTerraformTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terraform Tests Suite")
}

var (
	workingDir string

	googleCredentials = os.Getenv("GOOGLE_CREDENTIALS")
	googleProject     = os.Getenv("GOOGLE_PROJECT")
)

var _ = BeforeSuite(func() {
	workingDir = GinkgoT().TempDir()
	Expect(cp.Copy("../terraform", workingDir)).NotTo(HaveOccurred())
})

func buildVars(defaults, overrides map[string]any) map[string]any {
	result := map[string]any{}
	maps.Copy(result, defaults)
	maps.Copy(result, overrides)
	return result
}
