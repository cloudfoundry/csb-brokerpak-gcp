package upgrade_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/brokerpaks"
	"csbbrokerpakgcp/acceptance-tests/helpers/environment"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	fromVersion           string
	developmentBuildDir   string
	releasedBuildDir      string
	csbGCPReleaseDir      string
	cloudServiceBrokerDir string
	metadata              environment.GCPMetadata
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
	metadata = environment.ReadGCPMetadata()

	if releasedBuildDir == "" { // Released dir not specified, so we should download a brokerpak
		if fromVersion == "" { // Version not specified, so use latest
			fromVersion = brokerpaks.LatestVersion()
		}

		releasedBuildDir = brokerpaks.DownloadBrokerpak(fromVersion, brokerpaks.TargetDir(fromVersion))
	}

	preflight(releasedBuildDir)

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

	// The -db-name and -db-secret flags will be replaced when upgrading
	// from a broker app-based setup to a broker virtual machine-based setup.
	// We use ops files to replace these values, based on the original broker app-based configuration.
	// see `createVm` function acceptance-tests/helpers/brokers/create.go:34
	// We set them here to create our temporary manifest file with a secret that will be replaced.
	// The temporary manifest file is using when creating the broker VM.
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
		"-db-secret",
		"secret-will-be-replaced",
	)

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Run()).To(Succeed(), "failed to run boshifier - manifest creator")
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
