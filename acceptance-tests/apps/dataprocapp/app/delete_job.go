package app

import (
	"context"
	"dataprocapp/credentials"
	"log"
	"net/http"

	dataproc "cloud.google.com/go/dataproc/apiv1"
	"github.com/gorilla/mux"
	dataprocpb "google.golang.org/genproto/googleapis/cloud/dataproc/v1"
)

func handleDeleteJob(jobClient dataproc.JobControllerClient, creds credentials.DataprocCredentials) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling deletion of job")

		jobName, ok := mux.Vars(r)["job"]
		if !ok {
			fail(w, http.StatusBadRequest, "job missing.")
			return
		}

		jobReq := &dataprocpb.DeleteJobRequest{
			ProjectId: creds.ProjectID,
			Region:    creds.Region,
			JobId:     jobName,
		}

		ctx := context.Background()
		err := jobClient.DeleteJob(ctx, jobReq)
		if err != nil {
			fail(w, http.StatusFailedDependency, "error with request to delete job: %v", err)
			return
		}

		w.WriteHeader(http.StatusGone)
		log.Printf("Job deleted")
	}
}
