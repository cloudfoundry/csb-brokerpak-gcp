package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"storageapp/internal/credentials"
)

func App(creds credentials.StorageCredentials) http.HandlerFunc {
	client, _ := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(creds.Credentials)))

	return func(w http.ResponseWriter, r *http.Request) {
		fileName := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodGet:
			handleDownload(w, r, fileName, client, creds.BucketName)
		case http.MethodPut:
			handleUpload(w, r, fileName, client, creds.BucketName)
		case http.MethodDelete:
			handleDelete(w, r, fileName, client, creds.BucketName)
		default:
			fail(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}
	}
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
