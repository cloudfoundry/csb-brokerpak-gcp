package connector

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"mysqlapp/internal/keyvalue"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	customCaConfigName = "custom-ca"
)

type Connector struct {
	Host        string
	Database    string
	Username    string
	Password    string
	Port        int
	SSLRootCert string
	SSLKey      string
	SSLCert     string
}

type Option func(*Connector, *mysql.Config) error

func (c *Connector) Connect(opts ...Option) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = c.Host
	cfg.User = c.Username
	cfg.Passwd = c.Password
	cfg.DBName = c.Database

	if err := withDefaults(opts...)(c, cfg); err != nil {
		return nil, err
	}

	if cfg.TLSConfig == "true" {
		cfg.TLSConfig = customCaConfigName
		// Example of final URL: db, err := sql.Open("mysql", "user@tcp(localhost:3306)/test?tls=custom-ca")
		if err := registerCustomCA(c.SSLRootCert, c.SSLKey, c.SSLCert); err != nil {
			return nil, fmt.Errorf("failed to register custom certificate: %w", err)
		}
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		return db, fmt.Errorf("failed to verify the connection to the database is still alive")
	}

	_, err = db.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`,
		keyvalue.TableName,
		keyvalue.KeyColumn,
		keyvalue.ValueColumn,
	))
	if err != nil {
		log.Fatalf("failed to create test table: %s", err)
	}

	return db, nil
}

func withOptions(opts ...Option) Option {
	return func(conn *Connector, cfg *mysql.Config) error {
		for _, o := range opts {
			if err := o(conn, cfg); err != nil {
				return err
			}
		}
		return nil
	}
}

func withDefaults(opts ...Option) Option {
	return withOptions(append([]Option{WithTLS("true")}, opts...)...)
}

func WithTLS(tls string) Option {
	return func(_ *Connector, cfg *mysql.Config) error {
		switch tls {
		case "false", "true", "skip-verify", "preferred":
			cfg.TLSConfig = tls
		case "":
			cfg.TLSConfig = "true"
		default:
			return fmt.Errorf("invalid tls value: %s", tls)
		}
		return nil
	}
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
