package app

import (
	"fmt"
	"io"
	"log"
	"mysqlapp/internal/connector"
	"mysqlapp/internal/keyvalue"
	"net/http"
)

func handleSet(w http.ResponseWriter, r *http.Request, key string, conn connector.Connector) {
	log.Println("Handling set.")

	rawValue, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
		return
	}

	db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
	if err != nil {
		fail(w, http.StatusInternalServerError, "error connecting to database: %s", err)
	}

	stmt, err := db.Prepare(fmt.Sprintf(
		`INSERT INTO %s (%s, %s) VALUES (?, ?)`,
		keyvalue.TableName,
		keyvalue.KeyColumn,
		keyvalue.ValueColumn,
	))
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error preparing statement: %s", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(key, string(rawValue))
	if err != nil {
		fail(w, http.StatusBadRequest, "Error inserting values: %s", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Key %q set to value %q.", key, string(rawValue))
}
