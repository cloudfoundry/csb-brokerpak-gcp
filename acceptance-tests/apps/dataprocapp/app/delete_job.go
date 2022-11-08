package app

import (
	"context"
	"log"
	"net/http"

	dataproc "cloud.google.com/go/dataproc/apiv1"
	"cloud.google.com/go/dataproc/apiv1/dataprocpb"
	"github.com/gorilla/mux"

	"dataprocapp/credentials"
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
