package dataproc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDataproc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataproc Suite")
}
