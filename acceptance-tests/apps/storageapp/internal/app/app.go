package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"storageapp/internal/credentials"

	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
)

func App(creds credentials.StorageCredentials) *mux.Router {
	client, _ := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(creds.Credentials)))

	r := mux.NewRouter()

	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	r.HandleFunc("/{fileName}", handleUpload(client, creds.BucketName)).Methods("PUT")
	r.HandleFunc("/{fileName}", handleDownload(client, creds.BucketName)).Methods("GET")
	r.HandleFunc("/{fileName}", handleDelete(client, creds.BucketName)).Methods("DELETE")

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
