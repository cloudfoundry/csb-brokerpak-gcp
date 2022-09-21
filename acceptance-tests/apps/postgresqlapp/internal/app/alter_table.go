package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

func handleAlterTable(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling public.test table alteration.")

		_, err := db.Exec(fmt.Sprintf(`ALTER TABLE public.test alter column %s type varchar(256);`, valueColumn))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error altering the table %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Printf("table public.test modified")
	}
}
