// Package environment manages environment variables
package environment

import (
	"os"

	"github.com/onsi/gomega"
)

type GCPMetadata struct {
	Project     string
	Credentials string
}

func ReadGCPMetadata() GCPMetadata {
	result := GCPMetadata{
		Project:     os.Getenv("GOOGLE_PROJECT"),
		Credentials: os.Getenv("GOOGLE_CREDENTIALS"),
	}

	gomega.Expect(result.Project).NotTo(gomega.BeEmpty(), "must set GOOGLE_PROJECT")
	gomega.Expect(result.Credentials).NotTo(gomega.BeEmpty(), "must set GOOGLE_CREDENTIALS")

	return result
}
