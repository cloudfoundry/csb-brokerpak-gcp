package upgrade_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/brokerpaks"
	"csbbrokerpakgcp/acceptance-tests/helpers/environment"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	fromVersion         string
	developmentBuildDir string
	releasedBuildDir    string
	csbGCPRelease       string
	metadata            environment.GCPMetadata
)

func init() {
	flag.StringVar(&fromVersion, "from-version", "", "version to upgrade from")
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../", "location of development version of built broker and brokerpak")
	flag.StringVar(&csbGCPRelease, "csbGCPReleaseDir", "../../../csb-gcp-release", "location of development version of csb-gcp release")
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

	preflight(developmentBuildDir) // faster feedback as no download
	preflight(releasedBuildDir)

	absDevelopmentBuildDir, err := filepath.Abs(developmentBuildDir)
	Expect(err).NotTo(HaveOccurred())
	cmd := exec.Command(
		"go",
		"run",
		"-C",
		"../boshmanifestcreator/app/upgrader",
		".",
		"-brokerpak-path",
		absDevelopmentBuildDir,
		"-iaas-release-path",
		csbGCPRelease,
	)

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = os.Stderr

	Expect(cmd.Run()).To(Succeed(), "failed to run bosh manifest creator - upgrader")
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
