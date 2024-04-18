package app

import (
	"context"
	"dataprocapp/credentials"
	"fmt"
	"log"
	"net/http"
	"strings"

	dataproc "cloud.google.com/go/dataproc/v2/apiv1"
	"google.golang.org/api/option"
)

func App(creds credentials.DataprocCredentials) http.HandlerFunc {
	endpoint := fmt.Sprintf("%s-dataproc.googleapis.com:443", creds.Region)
	client, err := dataproc.NewJobControllerClient(
		context.Background(),
		option.WithEndpoint(endpoint),
		option.WithCredentialsJSON(creds.Credentials))
	if err != nil {
		log.Fatalf("error creating the cluster client: %s\n", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		job := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodGet:
			handleGetJob(w, job, client, creds)
		case http.MethodPut:
			handleRunJob(w, job, client, creds)
		case http.MethodDelete:
			handleDeleteJob(w, job, client, creds)
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
