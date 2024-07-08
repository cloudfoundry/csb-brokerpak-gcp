package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"mysqlapp/internal/app"
	"mysqlapp/internal/credentials"
)

func main() {
	log.Println("Starting.")

	log.Println("Reading credentials.")
	conn, err := credentials.Read()

	if err != nil {
		panic(err)
	}

	port := port()
	log.Printf("Listening on port: %s", port)
	http.Handle("/", app.App(conn))
	http.ListenAndServe(port, nil)
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
