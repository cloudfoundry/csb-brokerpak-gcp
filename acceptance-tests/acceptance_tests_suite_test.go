package acceptance_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"csbbrokerpakgcp/acceptance-tests/helpers/random"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	developmentBuildDir   string
	csbGCPReleaseDir      string
	cloudServiceBrokerDir string
)

func init() {
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../", "location of development version of built broker and brokerpak")
	flag.StringVar(&csbGCPReleaseDir, "csbGCPReleaseDir", "../../csb-gcp-release", "location of development version of csb-gcp release")
	flag.StringVar(&cloudServiceBrokerDir, "cloudServiceBrokerDir", "../../cloud-service-broker", "location of development version of cloud-service-broker release")
}

func TestAcceptanceTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Tests Suite")
}

var _ = BeforeSuite(func() {

	absDevelopmentBuildDir, err := filepath.Abs(developmentBuildDir)
	Expect(err).NotTo(HaveOccurred())

	absCSBGCPReleaseDir, err := filepath.Abs(csbGCPReleaseDir)
	Expect(err).NotTo(HaveOccurred())

	absCloudServiceBrokerDir, err := filepath.Abs(cloudServiceBrokerDir)
	Expect(err).NotTo(HaveOccurred())

	// We modify the release to use the local brokerpak, cloud-service-broker and iaas release
	// This is so that we can run the tests against the local brokerpak and cloud-service-broker
	// rather than the released versions. The command `vendir sync...` will modify the files, so we
	// prefer to run this in a temporary directory.
	tmpDir := os.TempDir()
	tmpReleasePath := fmt.Sprintf("%s/csb-gcp-release", tmpDir)
	GinkgoWriter.Printf("Running local release modifier - vendoring the brokerpak, cloud-service-broker and iaas release - destination %s\n", tmpReleasePath)

	cmd := exec.Command(
		"go",
		"run",
		"-C",
		"./boshifier/app/vendirlocalrelease",
		".",
		"-brokerpak-path",
		absDevelopmentBuildDir,
		"-cloud-service-broker-path",
		absCloudServiceBrokerDir,
		"-iaas-release-path",
		absCSBGCPReleaseDir,
		"-tmp-release-path",
		tmpReleasePath,
	)

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Run()).To(Succeed(), "failed to run boshifier - vendir local release")

	// The manifest creator requires a fixed secret to be passed in subsequent executions
	// to avoid mismatched secrets when encrypting and decrypting the database.
	const fixedSecret = "02630426-1d06-47b0-b712-5c74dd4f8182"
	cmd = exec.Command(
		"go",
		"run",
		"-C",
		"./boshifier/app/manifestcreator",
		".",
		"-brokerpak-path",
		absDevelopmentBuildDir,
		"-iaas-release-path",
		tmpReleasePath,
		"-db-name",
		random.Name(random.WithPrefix("db"), random.WithMaxLength(20), random.WithDelimiter("")),
		"-db-secret",
		fixedSecret,
	)

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Run()).To(Succeed(), "failed to run boshifier - manifest creator")

	cmd = exec.Command(
		"go",
		"run",
		"-C",
		"./boshifier/app/deployer",
		".",
		"-iaas-release-path",
		tmpReleasePath,
		"-bosh-deployment-name",
		"cloud-service-broker-gcp",
	)

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Start()).To(Succeed(), "failed to start boshifier - deployer")
	Expect(cmd.Wait()).To(Succeed(), "failed to run boshifier - deployer")
})

var _ = AfterSuite(func() {
	cmd := exec.Command(
		"go",
		"run",
		"-C",
		"./boshifier/app/deleter",
		".",
		"-bosh-deployment-name",
		"cloud-service-broker-gcp",
	)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Start()).To(Succeed(), "failed to start boshifier - deleter")
	Expect(cmd.Wait()).To(Succeed(), "failed to run boshifier - deleter")
})
