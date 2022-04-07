package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	URI         string `mapstructure:"uri"`
	SSLCert     string `mapstructure:"sslcert"`
	SSLKey      string `mapstructure:"sslkey"`
	SSLRootCert string `mapstructure:"sslrootcert"`
}

func Read() (*Config, error) {
	app, err := cfenv.Current()
	if err != nil {
		return nil, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("postgresql")
	if err != nil {
		return nil, fmt.Errorf("error reading PostgreSQL service details")
	}
	c := &Config{}

	if err := mapstructure.Decode(svs[0].Credentials, c); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if c.URI == "" && c.SSLKey == "" && c.SSLCert == "" && c.SSLRootCert == "" {
		return nil, fmt.Errorf("parsed credentials are not valid")
	}

	fmt.Printf("config:: %v\n", c)

	return c, nil
}
