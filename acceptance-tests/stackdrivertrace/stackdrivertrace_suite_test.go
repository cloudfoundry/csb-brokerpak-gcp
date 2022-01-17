package stackdrivertrace_test

import (
	"acceptancetests/helpers/environment"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var GCPMetadata environment.GCPMetadata

func TestStackdrivetrace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stackdrivetrace Suite")
}

var _ = BeforeSuite(func() {
	GCPMetadata = environment.ReadGCPMetadata()
})
