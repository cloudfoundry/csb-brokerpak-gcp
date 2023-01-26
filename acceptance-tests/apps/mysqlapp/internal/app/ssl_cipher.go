package app

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func handleGetSSLCipher(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")

		var res struct {
			VariableName string `sql:"Variable_name" json:"variable_name"`
			Value        string `sql:"Value" json:"value"`
		}
		err := db.QueryRow("SHOW STATUS LIKE 'Ssl_cipher'").Scan(&res.VariableName, &res.Value)
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
}
