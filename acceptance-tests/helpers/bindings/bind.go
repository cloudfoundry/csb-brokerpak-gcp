// Package bindings manages service bindings
package bindings

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
	"csbbrokerpakgcp/acceptance-tests/helpers/random"
)

type Binding struct {
	name                string
	serviceInstanceName string
	appName             string
}

func Bind(serviceInstanceName, appName string, params string) *Binding {
	name := random.Name()
	args := []string{
		"bind-service",
		appName,
		serviceInstanceName,
		"--binding-name",
		name,
	}

	if params != "" {
		args = append(args, "-c", params)
	}
	cf.Run(args...)
	return &Binding{
		name:                name,
		serviceInstanceName: serviceInstanceName,
		appName:             appName,
	}
}
