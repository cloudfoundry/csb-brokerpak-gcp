package app

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"

	"log"
	"net/http"
)

func handleDownload(client *storage.Client, bucketName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling download.")

		key, ok := mux.Vars(r)["fileName"]
		if !ok {
			fail(w, http.StatusBadRequest, "Filename missing.")
			return
		}

		reader, err := client.Bucket(bucketName).Object(key).NewReader(context.Background())
		if err != nil {
			fail(w, http.StatusFailedDependency, "getting reader for bucket object: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")

		if _, err := io.Copy(w, reader); err != nil {
			fail(w, http.StatusInternalServerError, "io.Copy: %v", err)
		}

		if err := reader.Close(); err != nil {
			fail(w, http.StatusFailedDependency, "Reader.Close: %v", err)
			return
		}

		log.Println("Download done.")
	}
}
