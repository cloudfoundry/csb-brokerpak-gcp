// Package brokers manages service brokers
package brokers

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/deployments"
)

type Broker struct {
	app       *apps.App
	depl      *deployments.Deployment
	Name      string
	username  string
	password  string
	dir       string
	secrets   []EncryptionSecret
	envExtras []apps.EnvVar
}

// Secrets returns the encryption secrets for the broker
// This function is a temporary workaround to allow the transition
// from app-based to VM-based brokers
// The Broker code will be removed after the first broker VM-based release.
func (b *Broker) Secrets() []EncryptionSecret {
	return b.secrets
}
