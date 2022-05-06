package credentials

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type config struct {
	URI         string `mapstructure:"uri"`
	SSLCert     string `mapstructure:"sslcert"`
	SSLKey      string `mapstructure:"sslkey"`
	SSLRootCert string `mapstructure:"sslrootcert"`
	TLS         bool   `mapstructure:"require_ssl"`
}

func Read() (string, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("postgresql")
	if err != nil {
		return "", fmt.Errorf("error reading PostgreSQL service details")
	}

	var c config
	if err := mapstructure.Decode(svs[0].Credentials, &c); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %+v: %w", svs[0].Credentials, err)
	}

	if c.URI == "" {
		return "", fmt.Errorf("missing URI: %+v", svs[0].Credentials)
	}

	switch c.TLS {
	case true:
		return tlsURI(c)
	default:
		return c.URI, nil
	}
}

func tlsURI(c config) (string, error) {
	sslrootcert, err := writeTmp("sslrootcert", c.SSLRootCert)
	if err != nil {
		return "", err
	}
	sslkey, err := writeTmp("sslkey", c.SSLKey)
	if err != nil {
		return "", err
	}
	sslcert, err := writeTmp("sslcert", c.SSLCert)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s?sslmode=verify-ca&sslrootcert=%s&sslkey=%s&sslcert=%s", c.URI, sslrootcert, sslkey, sslcert), nil
}

func writeTmp(moniker, data string) (string, error) {
	if data == "" {
		return "", fmt.Errorf("TLS configuration error: %q is empty", moniker)
	}

	fh, err := os.CreateTemp("", "")
	if err != nil {
		return "", fmt.Errorf("error opening temporary file: %w", err)
	}
	defer fh.Close()

	_, err = fh.Write([]byte(data))
	if err != nil {
		return "", fmt.Errorf("error writing to temporary file %q: %w", fh.Name(), err)
	}

	return fh.Name(), nil
}
