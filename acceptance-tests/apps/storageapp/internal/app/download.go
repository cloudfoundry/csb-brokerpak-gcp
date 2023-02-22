package app

import (
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/storage"
)

func handleDownload(w http.ResponseWriter, r *http.Request, fileName string, client *storage.Client, bucketName string) {
	log.Println("Handling download.")

	reader, err := client.Bucket(bucketName).Object(fileName).NewReader(r.Context())
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
