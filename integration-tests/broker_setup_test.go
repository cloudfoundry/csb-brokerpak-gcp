package integration_tests

import (
	"encoding/json"
	testframework "github.com/cloudfoundry-incubator/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var mockTerraform testframework.TerraformMock
var broker testframework.TestInstance

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
		"GOOGLE_CREDENTIALS="+ BrokerGCPCreds,
		"GOOGLE_PROJECT="+ BrokerGCPProject,
		`GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS=` + string(postgresPlansJson),
	})).NotTo(HaveOccurred())
})
