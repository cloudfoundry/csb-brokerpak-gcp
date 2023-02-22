package app

import (
	"log"
	"net/http"

	"cloud.google.com/go/storage"
)

func handleDelete(w http.ResponseWriter, r *http.Request, fileName string, client *storage.Client, bucketName string) {
	log.Println("Handling delete.")

	if err := client.Bucket(bucketName).Object(fileName).Delete(r.Context()); err != nil {
		fail(w, http.StatusFailedDependency, "Delete: %v", err)
		return
	}
	log.Println("Blob deleted.")

	w.WriteHeader(http.StatusGone)
}
