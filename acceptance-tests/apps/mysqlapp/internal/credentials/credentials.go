package credentials

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/multierr"
)

const (
	customCaConfigName          = "custom-ca"
	newBindingFormatFeatureFlag = "NEW_BINDING_FORMAT_FEATURE_FLAG"
	enabled                     = "ENABLED"
)

func Read() (string, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", fmt.Errorf("error reading app env: %w", err)
	}

	svs, err := app.Services.WithTag("mysql")
	if err != nil {
		return "", fmt.Errorf("error reading MySQL service details")
	}

	var m binding
	if err := mapstructure.Decode(svs[0].Credentials, &m); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %w", err)
	}

	m.isNewBindingFormat = os.Getenv(newBindingFormatFeatureFlag) == enabled

	if err := m.validate(); err != nil {
		return "", fmt.Errorf("parsed credentials are not valid %s", err.Error())
	}

	c := mysql.NewConfig()
	c.TLSConfig = "true"
	c.Net = "tcp"
	c.Addr = fmt.Sprintf("%s:%d", m.Host, m.Port)
	c.User = m.Username
	c.Passwd = m.Password
	c.DBName = m.Database

	if m.isNewBindingFormat {
		log.Println("registering custom CA")
		c.TLSConfig = "skip-verify"
		if err := registerCustomCA(m.SSLRootCert, m.SSLKey, m.SSLCert); err != nil {
			return "", fmt.Errorf("failed to register custom certificate %s", err.Error())
		}
	}

	return c.FormatDSN(), nil
}

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
		err = multierr.Append(err, fmt.Errorf("empty host"))
	}

	if b.Username == "" {
		err = multierr.Append(err, fmt.Errorf("empty username"))
	}

	if b.Password == "" {
		err = multierr.Append(err, fmt.Errorf("empty password"))
	}

	if b.Database == "" {
		err = multierr.Append(err, fmt.Errorf("empty database"))
	}

	if b.Port == 0 {
		err = multierr.Append(err, fmt.Errorf("invalid port"))
	}

	if !b.isNewBindingFormat {
		return err
	}

	if b.SSLCert == "" {
		err = multierr.Append(err, fmt.Errorf("empty cert"))
	}

	if b.SSLKey == "" {
		err = multierr.Append(err, fmt.Errorf("empty key"))
	}

	if b.SSLRootCert == "" {
		err = multierr.Append(err, fmt.Errorf("empty root cert"))
	}

	return err
}

func registerCustomCA(rootCert, key, cert string) error {
	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM([]byte(rootCert)); !ok {
		return fmt.Errorf("unable to append CA cert:\n[ %v ]", rootCert)
	}

	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return fmt.Errorf("unable to parse a public/private key pair from a pair of PEM encoded data CA cert:\n[ %v ]", rootCert)
	}

	clientCert = append(clientCert, certs)
	err = mysql.RegisterTLSConfig(customCaConfigName, &tls.Config{
		RootCAs:      rootCertPool,
		MinVersion:   tls.VersionTLS12,
		Certificates: clientCert,
	})
	if err != nil {
		return fmt.Errorf("unable to register custom-ca mysql config: %s", err.Error())
	}
	return nil
}
