// Package bindings manages service bindings
package bindings

import (
	"fmt"
	"strings"
	"time"

	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type Binding struct {
	name                string
	serviceInstanceName string
	appName             string
}

const (
	bindTimeout     = 3 * time.Minute
	maxBindAttempts = 3
)

func Bind(serviceInstanceName, appName string, params string) *Binding {
	var lastStderr string

	for attempt := 1; attempt <= maxBindAttempts; attempt++ {
		// Generate a new binding name for each attempt
		name := random.Name()
		args := []string{
			"bind-service",
			appName,
			serviceInstanceName,
			"--binding-name",
			name,
		}

		if params != "" {
			args = append(args, "-c", params)
		}

		session := cf.Start(args...)
		Eventually(session, bindTimeout).Should(gexec.Exit())

		if session.ExitCode() == 0 {
			return &Binding{
				name:                name,
				serviceInstanceName: serviceInstanceName,
				appName:             appName,
			}
		}

		lastStderr = string(session.Err.Contents())

		// Only retry on timeout errors
		if !strings.Contains(lastStderr, "timed out") {
			// Non-timeout error - fail immediately
			Fail("Bind failed: " + lastStderr)
		}

		if attempt < maxBindAttempts {
			GinkgoWriter.Printf("Bind timed out (attempt %d/%d), retrying with new binding name...\n", attempt, maxBindAttempts)
			// Brief pause before retry to allow any in-flight operations to settle
			time.Sleep(5 * time.Second)
		}
	}

	Fail(fmt.Sprintf("Bind failed after %d attempts due to timeout: %s", maxBindAttempts, lastStderr))
	return nil // unreachable, but needed for compilation
}
