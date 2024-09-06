// Package brokers manages service brokers
package brokers

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"csbbrokerpakgcp/acceptance-tests/helpers/deployments"
)

type Broker struct {
	app            *apps.App
	depl           *deployments.Deployment
	Name           string
	username       string
	password       string
	dir            string
	boshReleaseDir string
	secrets        []EncryptionSecret
	envExtras      []apps.EnvVar
	isVmBased      bool
}
