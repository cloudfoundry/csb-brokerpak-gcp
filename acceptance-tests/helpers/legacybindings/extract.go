// Package legacybindings handle bindings from the legacy GCP broker
package legacybindings

import (
	"github.com/mitchellh/mapstructure"
	. "github.com/onsi/gomega"
)

type LegacyBinding struct {
	InstanceName string `mapstructure:"instance_name"`
	DatabaseName string `mapstructure:"database_name"`
	Username     string `mapstructure:"Username"`
	Password     string `mapstructure:"Password"`
}

func ExtractLegacyBinding(data any) (result LegacyBinding) {
	Expect(mapstructure.Decode(data, &result)).To(Succeed())
	Expect(result.InstanceName).NotTo(BeEmpty())
	Expect(result.DatabaseName).NotTo(BeEmpty())
	Expect(result.Username).NotTo(BeEmpty())
	Expect(result.Password).NotTo(BeEmpty())
	return result
}
