package upgrade_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"csbbrokerpakgcp/acceptance-tests/helpers/brokerpaks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	fromVersion           string
	developmentBuildDir   string
	releasedBuildDir      string
	csbGCPReleaseDir      string
	cloudServiceBrokerDir string
)

func init() {
	flag.StringVar(&fromVersion, "from-version", "", "version to upgrade from")
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../", "location of development version of built broker and brokerpak")
	flag.StringVar(&csbGCPReleaseDir, "csbGCPReleaseDir", "../../../csb-gcp-release", "location of development version of csb-gcp release")
	flag.StringVar(&cloudServiceBrokerDir, "cloudServiceBrokerDir", "../../../cloud-service-broker", "location of development version of cloud-service-broker release")
}

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}

var _ = BeforeSuite(func() {
	Expect(os.Getenv("GOOGLE_PROJECT")).NotTo(BeEmpty(), "must set GOOGLE_PROJECT")
	Expect(os.Getenv("GOOGLE_CREDENTIALS")).NotTo(BeEmpty(), "must set GOOGLE_CREDENTIALS")

	if releasedBuildDir == "" { // Released dir not specified, so we should download a brokerpak
		if fromVersion == "" { // Version not specified, so use latest
			fromVersion = brokerpaks.LatestVersion()
		}

		releasedBuildDir = brokerpaks.DownloadBrokerpak(fromVersion, brokerpaks.TargetDir(fromVersion))
	}

	preflight(releasedBuildDir)
})

// preflight checks that a specified broker dir is viable so that the user gets fast feedback
func preflight(dir string) {
	GinkgoHelper()

	entries, err := os.ReadDir(dir)
	Expect(err).NotTo(HaveOccurred())
	names := make([]string, len(entries))
	for i := range entries {
		names[i] = entries[i].Name()
	}

	Expect(names).To(ContainElements(
		Equal("cloud-service-broker"),
		Equal(".envrc"),
		MatchRegexp(`gcp-services-\S+\.brokerpak`),
	))
}

func UpdateBrokerToVM(brokerName, brokerAppBasedSecret string) func() {
	absDevelopmentBuildDir, err := filepath.Abs(developmentBuildDir)
	Expect(err).NotTo(HaveOccurred())

	absCSBGCPReleaseDir, err := filepath.Abs(csbGCPReleaseDir)
	Expect(err).NotTo(HaveOccurred())

	absCloudServiceBrokerDir, err := filepath.Abs(cloudServiceBrokerDir)
	Expect(err).NotTo(HaveOccurred())

	// We modify the release to use the local brokerpak, cloud-service-broker and iaas release
	// This is so that we can run the tests against the local brokerpak and cloud-service-broker
	// rather than the released versions.  The command `vendir sync...` will modify the files, so we
	// prefer to run this in a temporary directory.
	tmpDir := os.TempDir()
	tmpReleasePath := fmt.Sprintf("%s/csb-gcp-release", tmpDir)
	GinkgoWriter.Printf("Running local release modifier - vendoring the brokerpak, cloud-service-broker and iaas release - destination %s\n", tmpReleasePath)

	cmd := exec.Command(
		"go",
		"run",
		"-C",
		"../boshifier/app/vendirlocalrelease",
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

	// The -db-name and -db-secret flags must be set to the values that were used in the original broker app-based setup.
	// The temporary manifest file is using when creating the broker VM.
	// We create the temp manifest file in the /tmp/tmp-manifest.yml directory to avoid committing it to the repository.
	cmd = exec.Command(
		"go",
		"run",
		"-C",
		"../boshifier/app/manifestcreator",
		".",
		"-brokerpak-path",
		absDevelopmentBuildDir,
		"-iaas-release-path",
		tmpReleasePath,
		"-db-name",
		strings.ReplaceAll(brokerName, "-", "_"),
		"-db-secret",
		brokerAppBasedSecret,
	)

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Run()).To(Succeed(), "failed to run boshifier - manifest creator")

	cmd = exec.Command(
		"go",
		"run",
		"-C",
		"../boshifier/app/deployer",
		".",
		"-iaas-release-path",
		tmpReleasePath,
		"-bosh-deployment-name",
		brokerName,
	)

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Start()).To(Succeed(), "failed to start boshifier - deployer")
	Expect(cmd.Wait()).To(Succeed(), "failed to run boshifier - deployer")

	return func() {
		cmd := exec.Command(
			"go",
			"run",
			"-C",
			"../boshifier/app/deleter",
			".",
			"-bosh-deployment-name",
			brokerName,
		)
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		Expect(cmd.Start()).To(Succeed(), "failed to start boshifier - deleter")
		Expect(cmd.Wait()).To(Succeed(), "failed to run boshifier - deleter")
	}
}
