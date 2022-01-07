package bindings

import (
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
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
