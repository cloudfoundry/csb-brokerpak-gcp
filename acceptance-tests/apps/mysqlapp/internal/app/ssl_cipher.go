package app

import (
	"encoding/json"
	"log"
	"mysqlapp/internal/connector"
	"net/http"
)

func handleGetSSLCipher(w http.ResponseWriter, r *http.Request, conn connector.Connector) {
	log.Println("Handling get.")

	var res struct {
		VariableName string `sql:"Variable_name" json:"variable_name"`
		Value        string `sql:"Value" json:"value"`
	}

	db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
	if err != nil {
		fail(w, http.StatusInternalServerError, "error connecting to database: %s", err)
	}

	err = db.QueryRow("SHOW STATUS LIKE 'Ssl_cipher'").Scan(&res.VariableName, &res.Value)
	if err != nil {
		fail(w, http.StatusNotFound, "Error executing query: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error writing value: %s", err)
		return
	}
}
