package spanner_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpanner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spanner Suite")
}
