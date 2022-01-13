package app

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

func handleUpload(client *storage.Client, bucketName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling upload.")

		key, ok := mux.Vars(r)["fileName"]
		if !ok {
			fail(w, http.StatusBadRequest, "Filename missing.")
			return
		}

		wc := client.Bucket(bucketName).Object(key).NewWriter(context.Background())
		if _, err := io.Copy(wc, r.Body); err != nil {
			fail(w, http.StatusFailedDependency, "io.Copy: %v", err)
			return
		}
		if err := wc.Close(); err != nil {
			fail(w, http.StatusFailedDependency, "Writer.Close: %v", err)
			return
		}
		log.Println("Blob uploaded.")

		w.WriteHeader(http.StatusCreated)
	}
}
