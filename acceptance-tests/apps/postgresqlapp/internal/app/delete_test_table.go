package app

import (
	"database/sql"
	"log"
	"net/http"
)

func handleDeleteTestTable(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling public.test table deletion.")

		_, err := db.Exec(`DROP TABLE public.test`)
		if err != nil {
			fail(w, http.StatusBadRequest, "Error dropping table public.test %v", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("table public.test dropped")
	}
}
