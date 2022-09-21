package credentials

import (
	"fmt"
	"net/url"
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
	svc, err := app.Services.WithTag("postgresql")

	legacySvc, legacyErr := app.Services.WithTag("postgres")

	if legacyErr == nil && svc == nil {
		fmt.Println("Using legacy binding")
		return legacyRead(legacySvc[0].Credentials["uri"].(string))
	}

	if err != nil {
		return "", fmt.Errorf("error reading PostgreSQL service details")
	}

	fmt.Println("Using csb binding")
	var c config

	if err := mapstructure.Decode(svc[0].Credentials, &c); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %+v: %w", svc[0].Credentials, err)
	}

	if c.URI == "" {
		return "", fmt.Errorf("missing URI: %+v", svc[0].Credentials)
	}

	switch c.TLS {
	case true:
		return tlsURI(c)
	default:
		return c.URI, nil
	}
}

func legacyRead(URI string) (string, error) {
	URL, err := url.Parse(URI)
	if err != nil {
		return "", fmt.Errorf("unable to parse db connection url %s: %s", URI, err.Error())
	}

	query := URL.Query()

	sslcert, err := writeTmp("sslcert", query.Get("sslcert"))
	if err != nil {
		return "", fmt.Errorf("unable to extract TLS certificate from connection uri: %s", err.Error())
	}
	query.Set("sslcert", sslcert)

	sslrootcert, err := writeTmp("sslrootcert", query.Get("sslrootcert"))
	if err != nil {
		return "", fmt.Errorf("unable to extract TLS CA certificate from connection uri: %s", err.Error())
	}
	query.Set("sslrootcert", sslrootcert)

	sslkey, err := writeTmp("sslkey", query.Get("sslkey"))
	if err != nil {
		return "", fmt.Errorf("unable to extract TLS key from connection uri: %s", err.Error())
	}
	query.Set("sslkey", sslkey)
	URL.RawQuery = query.Encode()

	return URL.String(), nil
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
