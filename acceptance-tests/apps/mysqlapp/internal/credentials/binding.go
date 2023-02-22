package credentials

import (
	"errors"
	"fmt"
)

type binding struct {
	Host               string `mapstructure:"hostname"`
	Database           string `mapstructure:"name"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	Port               int    `mapstructure:"port"`
	SSLCert            string `mapstructure:"sslcert"`
	SSLKey             string `mapstructure:"sslkey"`
	SSLRootCert        string `mapstructure:"sslrootcert"`
	isNewBindingFormat bool
}

func (b binding) validate() error {
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

	if b.Port == 0 {
		err = errors.Join(err, fmt.Errorf("invalid port"))
	}

	if !b.isNewBindingFormat {
		return err
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
