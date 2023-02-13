package apps

import (
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
)

func (a *App) Restart() {
	Restart(a)
}

func Restart(apps ...*App) {
	for _, app := range apps {
		session := cf.Start("restart", app.Name)
		Eventually(session, 5*time.Minute).Should(gexec.Exit())
		checkSuccess(session.ExitCode(), app.Name)
	}
}
