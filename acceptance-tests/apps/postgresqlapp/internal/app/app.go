package app

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(uri string) *mux.Router {
	db := connect(uri)

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods(http.MethodHead, http.MethodGet)
	r.HandleFunc("/{key}", handleSet(db)).Methods(http.MethodPut)
	r.HandleFunc("/{key}", handleGet(db)).Methods(http.MethodGet)

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(uri string) *sql.DB {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	db.SetMaxIdleConns(0)

	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS public.%s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, tableName, keyColumn, valueColumn))
	if err != nil {
		log.Fatalf("Error creating table: %s", err)
	}

	_, err = db.Exec(fmt.Sprintf(`GRANT ALL ON TABLE public.%s TO PUBLIC`, tableName))
	if err != nil {
		log.Fatalf("Error granting table permissions: %s", err)
	}

	return db
}

func fail(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
