package app

import (
	"context"
	"log"
	"net/http"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/spanner"
	"github.com/gorilla/mux"
)

func handleGet(client spanner.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")

		key, ok := mux.Vars(r)["key"]
		if !ok {
			fail(w, http.StatusBadRequest, "Key missing.")
			return
		}

		stmt := spanner.Statement{SQL: `SELECT valuedata FROM test WHERE keyname = '` + key + "'"}
		iter := client.Single().Query(context.Background(), stmt)
		defer iter.Stop()
		row, err := iter.Next()
		if err == iterator.Done {
			fail(w, http.StatusNotFound, "key not found: %s", key)
			return
		}
		if err != nil {
			fail(w, http.StatusFailedDependency, "error querying database: %s", err)
			return

		}
		var valuedata string
		if err := row.Columns(&valuedata); err != nil {
			fail(w, http.StatusNotFound, "key not found: %s", key)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write([]byte(valuedata))

		if err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}
		log.Printf("Value %q retrived from key %q.", valuedata, key)
	}
}
