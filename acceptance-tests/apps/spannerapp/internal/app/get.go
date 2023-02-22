package app

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/spanner"
)

func handleGet(w http.ResponseWriter, key string, client *spanner.Client) {
	log.Println("Handling get.")

	row, err := client.Single().ReadRow(context.Background(), tableName, spanner.Key{key}, []string{valueColumn})
	if err != nil {
		log.Printf("Error reading row: %s", err)
		return
	}

	var valuedata string

	if err := row.Columns(&valuedata); err != nil {
		fail(w, http.StatusFailedDependency, "could not read value for key '%s': %s", key, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(valuedata))

	if err != nil {
		log.Printf("Error writing value: %s", err)
		return
	}
	log.Printf("Value %q retrieved from key %q.", valuedata, key)
}
