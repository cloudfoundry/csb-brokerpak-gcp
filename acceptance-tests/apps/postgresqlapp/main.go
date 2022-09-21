package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"postgresqlapp/internal/app"
	"postgresqlapp/internal/credentials"
)

func main() {
	log.Println("Starting.")

	log.Println("Reading credentials.")
	uri, err := credentials.Read()
	if err != nil {
		panic(err)
	}

	port := port()
	log.Printf("Listening on port: %s", port)
	http.Handle("/", app.App(uri))
	log.Printf("Handlers init completed")
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Panic(err)
	}
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
