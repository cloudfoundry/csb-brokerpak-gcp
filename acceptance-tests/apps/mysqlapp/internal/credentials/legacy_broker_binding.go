package credentials

import (
	"errors"
	"fmt"
)

type legacyBrokerBinding struct {
	Host        string `mapstructure:"host"`
	Database    string `mapstructure:"database_name"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"Password"`
	SSLRootCert string `mapstructure:"CaCert"`
	SSLKey      string `mapstructure:"ClientKey"`
	SSLCert     string `mapstructure:"ClientCert"`
}

func (b legacyBrokerBinding) validate() error {
	var err error

	if b.Host == "" {
		err = errors.Join(err, fmt.Errorf("empty host"))
	}

	if b.Username == "" {
		err = errors.Join(err, fmt.Errorf("empty username"))
	}

	if b.Password == "" {
		err = errors.Join(err, fmt.Errorf("empty password"))
	}

	if b.Database == "" {
		err = errors.Join(err, fmt.Errorf("empty database"))
	}

	if b.SSLCert == "" {
		err = errors.Join(err, fmt.Errorf("empty cert"))
	}

	if b.SSLKey == "" {
		err = errors.Join(err, fmt.Errorf("empty key"))
	}

	if b.SSLRootCert == "" {
		err = errors.Join(err, fmt.Errorf("empty root cert"))
	}

	return err
}
