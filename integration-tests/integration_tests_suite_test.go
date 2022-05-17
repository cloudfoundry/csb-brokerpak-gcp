package integration_test

import (
	"encoding/json"
	"testing"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTests Suite")
}

var (
	mockTerraform testframework.TerraformMock
	broker        *testframework.TestInstance
)

const (
	brokerGCPProject = "broker-gcp-project"
	brokerGCPCreds   = "broker-gcp-creds"
)

var _ = BeforeSuite(func() {
	var err error
	mockTerraform, err = testframework.NewTerraformMock()
	Expect(err).NotTo(HaveOccurred())

	broker, err = testframework.BuildTestInstance(testframework.PathToBrokerPack(), mockTerraform, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	Expect(broker.Start(GinkgoWriter, []string{
		"GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true",
		"GOOGLE_CREDENTIALS=" + brokerGCPCreds,
		"GOOGLE_PROJECT=" + brokerGCPProject,
		`GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS=` + marshall(postgresPlans),
		"GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS=" + marshall(customMySQLPlans),
		"CSB_LISTENER_HOST=localhost", // prevents permissions popup
	})).To(Succeed())
})

var _ = AfterSuite(func() {
	if broker != nil {
		Expect(broker.Cleanup()).To(Succeed())
	}
})

func marshall(element any) string {
	b, err := json.Marshal(element)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}
