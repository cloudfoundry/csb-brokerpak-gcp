package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(uri string) http.Handler {
	db := connect(uri)

	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && strings.Trim(r.URL.Path, "/") == "":
			aliveness(w, r)
		default:
			methodNotAllowed(w)
		}
	})

	r.HandleFunc("/key-value/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodGet:
			handleGet(w, key, db)
		case http.MethodPut:
			handleSet(w, r, key, db)
		default:
			methodNotAllowed(w)
		}
	})

	r.HandleFunc("/admin/ssl/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			handleGetSSLCipher(w, db)
		default:
			methodNotAllowed(w)
		}
	})

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(uri string) *sql.DB {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, tableName, keyColumn, valueColumn))
	if err != nil {
		log.Fatalf("failed to create test table: %s", err)
	}

	return db
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}

func methodNotAllowed(w http.ResponseWriter) {
	fail(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
}
