package app

import (
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/storage"
)

func handleUpload(w http.ResponseWriter, r *http.Request, fileName string, client *storage.Client, bucketName string) {
	log.Println("Handling upload.")

	wc := client.Bucket(bucketName).Object(fileName).NewWriter(r.Context())
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
