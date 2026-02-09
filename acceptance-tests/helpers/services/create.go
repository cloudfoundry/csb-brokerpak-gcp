// Package services manages service instances
package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"csbbrokerpakgcp/acceptance-tests/helpers/brokers"
	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type ServiceInstance struct {
	Name string
	guid string
}

type config struct {
	name              string
	serviceBrokerName func() string
	parameters        string
	plan              string
}

type Option func(*config)

func CreateInstance(offering, plan string, opts ...Option) *ServiceInstance {
	cfg := defaultConfig(offering, plan, opts...)
	brokerName := cfg.serviceBrokerName()
	args := []string{
		"create-service",
		"--wait",
		offering,
		plan,
		cfg.name,
		"-b",
		brokerName,
	}

	if cfg.parameters != "" {
		args = append(args, "-c", cfg.parameters)
	}

	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		return gatherFailureDiagnostics(cfg.name, brokerName)
	})

	return &ServiceInstance{Name: cfg.name}
}

// gatherFailureDiagnostics collects diagnostic information when service creation fails
func gatherFailureDiagnostics(serviceName, brokerName string) string {
	var diagnostics strings.Builder

	diagnostics.WriteString("\n========== SERVICE INSTANCE STATUS ==========\n")
	serviceOut, _ := cf.Run("service", serviceName)
	diagnostics.WriteString(serviceOut)

	// Get broker app logs which contain the full Terraform error output
	diagnostics.WriteString("\n========== BROKER APP LOGS (recent) ==========\n")
	GinkgoWriter.Printf("Fetching broker logs for: %s\n", brokerName)

	logsSession := cf.Start("logs", brokerName, "--recent")
	Eventually(logsSession, 2*time.Minute).Should(Exit())
	if logsSession.ExitCode() == 0 {
		logsOutput := string(logsSession.Out.Contents())
		lines := strings.Split(logsOutput, "\n")
		if len(lines) > 200 {
			diagnostics.WriteString(fmt.Sprintf("... (showing last 200 of %d lines)\n", len(lines)))
			lines = lines[len(lines)-200:]
		}
		diagnostics.WriteString(strings.Join(lines, "\n"))
	} else {
		diagnostics.WriteString(fmt.Sprintf("Failed to fetch broker logs: %s\n", string(logsSession.Err.Contents())))
	}

	diagnostics.WriteString("\n========== END DIAGNOSTICS ==========\n")
	return diagnostics.String()
}

func WithDefaultBroker() Option {
	return func(c *config) {
		c.serviceBrokerName = brokers.DefaultBrokerName
	}
}

func WithBroker(broker *brokers.Broker) Option {
	return func(c *config) {
		c.serviceBrokerName = func() string { return broker.Name }
	}
}

func WithParameters(parameters any) Option {
	return func(c *config) {
		switch p := parameters.(type) {
		case string:
			c.parameters = p
		default:
			params, err := json.Marshal(p)
			Expect(err).NotTo(HaveOccurred())
			c.parameters = string(params)
		}
	}
}

func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

func WithOptions(opts ...Option) Option {
	return func(c *config) {
		for _, o := range opts {
			o(c)
		}
	}
}

func defaultConfig(offering, plan string, opts ...Option) config {
	var cfg config
	WithOptions(append([]Option{
		WithDefaultBroker(),
		WithName(random.Name(random.WithPrefix(offering, plan))),
	}, opts...)...)(&cfg)
	return cfg
}
