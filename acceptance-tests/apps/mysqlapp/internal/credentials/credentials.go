package credentials

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
)

const customCaConfigName = "custom-ca"

func Read() (string, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", fmt.Errorf("error reading app env: %w", err)
	}

	svs, err := app.Services.WithTag("mysql")
	if err != nil {
		return "", fmt.Errorf("error reading MySQL service details")
	}

	switch svs[0].Label {
	case "google-cloudsql-mysql-vpc":
		return readLegacyBrokerBinding(svs[0].Credentials)
	default:
		return readBinding(svs[0].Credentials)
	}
}

func readBinding(creds any) (string, error) {
	var m binding
	if err := mapstructure.Decode(creds, &m); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %w", err)
	}

	if err := m.validate(); err != nil {
		return "", fmt.Errorf("parsed credentials are not valid: %w", err)
	}

	c := mysql.NewConfig()
	c.TLSConfig = "false"
	c.Net = "tcp"
	c.Addr = fmt.Sprintf("%s:%d", m.Host, m.Port)
	c.User = m.Username
	c.Passwd = m.Password
	c.DBName = m.Database

	c.TLSConfig = customCaConfigName
	if err := registerCustomCA(m.SSLRootCert, m.SSLKey, m.SSLCert); err != nil {
		return "", fmt.Errorf("failed to register custom certificate: %w", err)
	}

	return c.FormatDSN(), nil
}

func readLegacyBrokerBinding(creds any) (string, error) {
	var m legacyBrokerBinding
	if err := mapstructure.Decode(creds, &m); err != nil {
		return "", err
	}

	if err := m.validate(); err != nil {
		return "", fmt.Errorf("parsed legacy broker credentials are not valid: %w", err)
	}

	c := mysql.NewConfig()
	c.TLSConfig = "false"
	c.Net = "tcp"
	c.Addr = fmt.Sprintf("%s:3306", m.Host)
	c.User = m.Username
	c.Passwd = m.Password
	c.DBName = m.Database

	c.TLSConfig = customCaConfigName
	if err := registerCustomCA(m.SSLRootCert, m.SSLKey, m.SSLCert); err != nil {
		return "", fmt.Errorf("failed to register custom certificate: %w", err)
	}

	return c.FormatDSN(), nil
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
	log.Printf("registering custom cert")
	err = mysql.RegisterTLSConfig(customCaConfigName, &tls.Config{
		RootCAs:            rootCertPool,
		MinVersion:         tls.VersionTLS12,
		Certificates:       clientCert,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return fmt.Errorf("unable to register custom-ca mysql config: %s", err.Error())
	}
	return nil
}
