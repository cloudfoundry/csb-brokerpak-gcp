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
	BrokerGCPProject = "broker-gcp-project"
	BrokerGCPCreds   = "broker-gcp-creds"
)

var _ = BeforeSuite(func() {
	var err error
	mockTerraform, err = testframework.NewTerraformMock()
	Expect(err).ShouldNot(HaveOccurred())

	broker, err = testframework.BuildTestInstance(testframework.PathToBrokerPack(), mockTerraform, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())

	postgresPlansJSON, err := json.Marshal(postgresPlans)
	Expect(err).ShouldNot(HaveOccurred())

	mySQLPlansJSON, err := json.Marshal([]map[string]interface{}{customMySQLPlan})
	Expect(err).ShouldNot(HaveOccurred())

	Expect(broker.Start(GinkgoWriter, []string{
		"GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true",
		"GOOGLE_CREDENTIALS=" + BrokerGCPCreds,
		"GOOGLE_PROJECT=" + BrokerGCPProject,
		`GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS=` + string(postgresPlansJSON),
		"GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS=" + string(mySQLPlansJSON),
		"CSB_LISTENER_HOST=localhost", // prevents permissions popup
	})).To(Succeed())
})

var _ = AfterSuite(func() {
	if broker != nil {
		Expect(broker.Cleanup()).To(Succeed())
	}
})
