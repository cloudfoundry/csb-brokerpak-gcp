package app

import (
	"context"
	"dataprocapp/credentials"
	"fmt"
	"log"
	"net/http"

	dataproc "cloud.google.com/go/dataproc/apiv1"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

func App(creds credentials.DataprocCredentials) *mux.Router {
	endpoint := fmt.Sprintf("%s-dataproc.googleapis.com:443", creds.Region)
	client, err := dataproc.NewJobControllerClient(
		context.Background(),
		option.WithEndpoint(endpoint),
		option.WithCredentialsJSON(creds.Credentials))
	if err != nil {
		log.Fatalf("error creating the cluster client: %s\n", err)
	}
	r := mux.NewRouter()

	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	r.HandleFunc("/{job}", handleRunJob(*client, creds)).Methods("PUT")
	r.HandleFunc("/{job}", handleGetJob(*client, creds)).Methods("GET")
	r.HandleFunc("/{job}", handleDeleteJob(*client, creds)).Methods("DELETE")
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
