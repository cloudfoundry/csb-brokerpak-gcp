package integration_tests

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

var mockTerraform testframework.TerraformMock
var broker *testframework.TestInstance

var BrokerGCPProject = "broker-gcp-project"
var BrokerGCPCreds = "broker-gcp-creds"

var _ = BeforeSuite(func() {
	var err error
	mockTerraform, err = testframework.NewTerraformMock()
	Expect(err).NotTo(HaveOccurred())

	broker, err = testframework.BuildTestInstance(testframework.PathToBrokerPack(), mockTerraform, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	postgresPlansJson, err := json.Marshal(postgresPlans)
	Expect(err).NotTo(HaveOccurred())

	Expect(broker.Start(GinkgoWriter, []string{
		"GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true",
		"GOOGLE_CREDENTIALS=" + BrokerGCPCreds,
		"GOOGLE_PROJECT=" + BrokerGCPProject,
		`GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS=` + string(postgresPlansJson),
	})).To(Succeed())
})

var _ = AfterSuite(func() {
	if broker != nil {
		Expect(broker.Cleanup()).To(Succeed())
	}
})
