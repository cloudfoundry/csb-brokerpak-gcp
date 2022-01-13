package cf

import (
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func Run(args ...string) (string, string) {
	session := Start(args...)
	Eventually(session, 3*time.Minute).Should(gexec.Exit(0))
	return string(session.Out.Contents()), string(session.Err.Contents())
}
