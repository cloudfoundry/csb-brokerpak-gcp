package csbpg

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type connectionFactory struct {
	host          string
	port          int
	username      string
	password      string
	database      string
	dataOwnerRole string
	sslClientCert *clientCertificateConfig
	sslRootCert   string
	sslMode       string
}

func (c connectionFactory) Connect() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.uri())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL %q: %w", c.uriRedacted(), err)
	}

	return db, nil
}

func (c connectionFactory) uri() string {
	return strings.Join([]string{
		"host=" + c.host,
		fmt.Sprintf("port=%d", c.port),
		"user=" + c.username,
		"password=" + c.password,
		"database=" + c.database,
		"sslmode=" + c.sslMode,
		"sslinline=true",
		fmt.Sprintf("sslcert='%s'", c.sslClientCert.Certificate),
		fmt.Sprintf("sslkey='%s'", c.sslClientCert.Key),
		fmt.Sprintf("sslrootcert='%s'", c.sslRootCert),
	}, " ")
}

func (c connectionFactory) uriRedacted() string {
	return strings.ReplaceAll(c.uri(), c.password, "REDACTED")
}

type clientCertificateConfig struct {
	Certificate string
	Key         string
}
