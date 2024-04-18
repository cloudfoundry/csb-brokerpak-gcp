package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(uri string) http.Handler {
	db, err := connect(uri)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connection succeeded")

	r := chi.NewRouter()
	r.Head("/", aliveness)
	r.Put("/{schema}", handleCreateSchema(db))
	r.Delete("/{schema}", handleDropSchema(db))

	// Although the URL path implies that these might do something in the schema, in fact
	// they ignore the schema name and just use the public schema
	r.Put("/{schema}/{key}", handleSet(db))
	r.Get("/{schema}/{key}", handleGet(db))

	// Although this takes a schema and table name as parameters, it ignores them
	r.Put("/schemas/{schema}/{table}", handleAlterTable(db))

	// This should be moved to a more meaningful URL path
	r.Delete("/", handleDeleteTestTable(db))

	return r
}

func aliveness(w http.ResponseWriter, _ *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(uri string) (*sql.DB, error) {
	db, err := sql.Open("pgx", uri)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to connect to database", err)
	}
	db.SetMaxIdleConns(0)

	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS public.%s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, tableName, keyColumn, valueColumn))
	if err != nil {
		return nil, fmt.Errorf("%w: error creating table", err)
	}

	_, err = db.Exec(fmt.Sprintf(`GRANT ALL ON TABLE public.%s TO PUBLIC`, tableName))
	if err != nil {
		return nil, fmt.Errorf("%w: error granting table permissions", err)
	}

	return db, nil
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}

func schemaName(r *http.Request) (string, error) {
	schema := chi.URLParam(r, "schema")

	switch {
	case schema == "":
		return "", fmt.Errorf("schema not specified or empty")
	case len(schema) > 50:
		return "", fmt.Errorf("schema name too long")
	case !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(schema):
		return "", fmt.Errorf("schema name contains invalid characters")
	default:
		return schema, nil
	}
}
