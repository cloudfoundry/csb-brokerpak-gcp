package app

import (
	"context"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/spanner"
)

func handleSet(w http.ResponseWriter, r *http.Request, key string, client *spanner.Client) {
	log.Println("Handling set.")

	rawValue, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
		return
	}

	columns := []string{keyColumn, valueColumn}

	m := []*spanner.Mutation{
		spanner.InsertOrUpdate(tableName, columns, []any{key, string(rawValue)}),
	}
	_, err = client.Apply(context.Background(), m)
	if err != nil {
		fail(w, http.StatusFailedDependency, "Error inserting data: %s", err)
		return
	}

	log.Printf("Key %q set to value %q.", key, string(rawValue))
}
