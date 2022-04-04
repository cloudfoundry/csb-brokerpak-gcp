package acceptance_test

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/environment"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAcceptanceTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Tests Suite")
}

var GCPMetadata environment.GCPMetadata

var _ = BeforeSuite(func() {
	GCPMetadata = environment.ReadGCPMetadata()
})
