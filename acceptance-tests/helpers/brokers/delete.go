package brokers

import (
	"os/exec"
	"time"

	"csbbrokerpakgcp/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
)

func (b *Broker) Delete() {
	if b.isVmBased {
		b.deleteDeployment()
	} else {
		b.deleteApp()
	}
}

func (b *Broker) deleteDeployment() {
	deployCmd := exec.Command(
		"bosh",
		"-n", "delete-deployment", "-d", b.Name,
	)
	session, err := gexec.Start(deployCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).To(Not(HaveOccurred()))
	Eventually(session, time.Minute*30).Should(gexec.Exit(0))
}

func (b *Broker) deleteApp() {
	cf.Run("delete-service-broker", b.Name, "-f")
	b.app.Delete()
}
