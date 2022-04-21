package csbpg

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const uriFormat = "postgres://%s:%s@%s:%d/%s?sslmode=disable"

type connectionFactory struct {
	host          string
	port          int
	username      string
	password      string
	database      string
	dataOwnerRole string
}

func (c connectionFactory) Connect() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.uri())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL %q: %w", c.uriRedacted(), err)
	}

	return db, nil
}

func (c connectionFactory) uri() string {
	return fmt.Sprintf(uriFormat, c.username, c.password, c.host, c.port, c.database)
}

func (c connectionFactory) uriRedacted() string {
	return fmt.Sprintf(uriFormat, c.username, "REDACTED", c.host, c.port, c.database)
}
