package services

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
)

func (s *ServiceInstance) Delete() {
	Delete(s.Name)
}

func Delete(name string) {
	switch cf.Version() {
	case cf.VersionV8:
		deleteWithWait(name)
	default:
		deleteWithPoll(name)
	}
}

func deleteWithWait(name string) {
	session := cf.Start("delete-service", "-f", name, "--wait")
	Eventually(session, time.Hour).Should(Exit(0))
}

func deleteWithPoll(name string) {
	cf.Run("delete-service", "-f", name)

	Eventually(func() string {
		out, _ := cf.Run("services")
		return out
	}, time.Hour, 30*time.Second).ShouldNot(ContainSubstring(name))
}
