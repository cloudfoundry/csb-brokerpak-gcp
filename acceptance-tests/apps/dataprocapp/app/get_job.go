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

func handleGetJob(jobClient dataproc.JobControllerClient, creds credentials.DataprocCredentials) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling getting job")

		jobName, ok := mux.Vars(r)["job"]
		if !ok {
			fail(w, http.StatusBadRequest, "job missing.")
			return
		}

		jobReq := &dataprocpb.GetJobRequest{
			ProjectId: creds.ProjectID,
			Region:    creds.Region,
			JobId:     jobName,
		}

		ctx := context.Background()
		getJobOp, err := jobClient.GetJob(ctx, jobReq)
		if err != nil {
			fail(w, http.StatusFailedDependency, "error with request to getting job: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write([]byte((getJobOp.Status.State.String())))

		log.Printf("Job finished with status: %s", getJobOp.Status.State.String())
	}
}
