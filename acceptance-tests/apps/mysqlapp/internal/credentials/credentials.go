package credentials

import (
	"fmt"
	"mysqlapp/internal/connector"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

func Read() (connector.Connector, error) {
	app, err := cfenv.Current()
	if err != nil {
		return connector.Connector{}, fmt.Errorf("error reading app env: %w", err)
	}

	svs, err := app.Services.WithTag("mysql")
	if err != nil {
		return connector.Connector{}, fmt.Errorf("error reading MySQL service details")
	}

	serviceHandlers := map[string]func(any) (connector.Connector, error){
		"google-cloudsql-mysql-vpc": readLegacyBrokerBinding,
	}

	// Attempt to find a handler for the service label.
	if handler, ok := serviceHandlers[svs[0].Label]; ok {
		return handler(svs[0].Credentials)
	}

	// Default to readBinding if no specific handler is found.
	return readBinding(svs[0].Credentials)
}

func readBinding(creds any) (connector.Connector, error) {
	var m binding
	if err := mapstructure.Decode(creds, &m); err != nil {
		return connector.Connector{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if err := m.validate(); err != nil {
		return connector.Connector{}, fmt.Errorf("parsed credentials are not valid: %w", err)
	}

	c := connector.Connector{
		Host:        fmt.Sprintf("%s:%d", m.Host, m.Port),
		Database:    m.Database,
		Username:    m.Username,
		Password:    m.Password,
		Port:        m.Port,
		SSLRootCert: m.SSLRootCert,
		SSLKey:      m.SSLKey,
		SSLCert:     m.SSLCert,
	}

	return c, nil
}

func readLegacyBrokerBinding(creds any) (connector.Connector, error) {
	var m legacyBrokerBinding
	if err := mapstructure.Decode(creds, &m); err != nil {
		return connector.Connector{}, err
	}

	if err := m.validate(); err != nil {
		return connector.Connector{}, fmt.Errorf("parsed legacy broker credentials are not valid: %w", err)
	}

	c := connector.Connector{
		Host:        fmt.Sprintf("%s:3306", m.Host),
		Database:    m.Database,
		Username:    m.Username,
		Password:    m.Password,
		Port:        3306,
		SSLRootCert: m.SSLRootCert,
		SSLKey:      m.SSLKey,
		SSLCert:     m.SSLCert,
	}

	return c, nil
}
