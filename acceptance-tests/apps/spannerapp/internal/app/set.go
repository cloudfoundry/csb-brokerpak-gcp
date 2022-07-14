package app

import (
	"context"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/spanner"
	"github.com/gorilla/mux"
)

func handleSet(client spanner.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		key, ok := mux.Vars(r)["key"]
		if !ok {
			fail(w, http.StatusBadRequest, "Key missing.")
			return
		}

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
}
