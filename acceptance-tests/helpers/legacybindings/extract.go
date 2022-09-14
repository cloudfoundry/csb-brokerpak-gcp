package legacybindings

import (
	"fmt"

	. "github.com/onsi/gomega"
)

type LegacyPostgresBinding struct {
	InstanceName string
	DatabaseName string
	Username     string
	Password     string
}

const (
	pgInstanceNameKey = "instance_name"
	pgDatabaseNameKey = "database_name"
	pgUsernameKey     = "Username"
	pgPasswordKey     = "Password"
)

func ExtractPostgresBinding(data any) (*LegacyPostgresBinding, error) {
	Expect(data).To(BeAssignableToTypeOf(map[string]any{}))
	bindingMap := data.(map[string]any)
	instanceName, err := extractPgBindingValue(bindingMap, pgInstanceNameKey)
	if err != nil {
		return nil, err
	}

	dbName, err := extractPgBindingValue(bindingMap, pgDatabaseNameKey)
	if err != nil {
		return nil, err
	}

	username, err := extractPgBindingValue(bindingMap, pgUsernameKey)
	if err != nil {
		return nil, err
	}

	password, err := extractPgBindingValue(bindingMap, pgPasswordKey)
	if err != nil {
		return nil, err
	}

	return &LegacyPostgresBinding{
		InstanceName: instanceName,
		DatabaseName: dbName,
		Username:     username,
		Password:     password,
	}, nil
}

func extractPgBindingValue(binding map[string]any, valueName string) (string, error) {
	value, ok := binding[valueName]
	if !ok {
		return "", fmt.Errorf("%s not found in binding data", valueName)
	}
	result, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected %#v to be a string", value)
	}
	return result, nil
}
