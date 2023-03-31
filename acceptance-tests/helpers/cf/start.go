package cf

import (
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func Start(args ...string) *gexec.Session {
	return start(nil, args...)
}

func StartWithWorkingDirectory(wd string, args ...string) *gexec.Session {
	return start(&wd, args...)
}

func start(pwd *string, args ...string) *gexec.Session {
	GinkgoWriter.Printf("Running: cf %s\n", strings.Join(args, " "))
	command := exec.Command("cf", args...)
	if pwd != nil {
		command.Dir = *pwd
	}
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
