// Package environment manages environment variables
package gcloud

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func GCP(args ...string) []byte {
	cmd := exec.Command("gcloud", args...)
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "Running: %s\n", cmd.String())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(session).WithTimeout(time.Minute).Should(gexec.Exit(0))
	return session.Out.Contents()
}

func GSUtil(args ...string) []byte {
	cmd := exec.Command("gsutil", args...)
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "Running: %s\n", cmd.String())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(session).WithTimeout(time.Minute).Should(gexec.Exit(0))
	return session.Out.Contents()
}
