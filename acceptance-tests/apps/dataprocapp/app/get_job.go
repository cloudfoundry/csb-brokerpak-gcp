package app

import (
	"context"
	"dataprocapp/credentials"
	"log"
	"net/http"

	dataproc "cloud.google.com/go/dataproc/v2/apiv1"
	"cloud.google.com/go/dataproc/v2/apiv1/dataprocpb"
)

func handleGetJob(w http.ResponseWriter, jobName string, jobClient *dataproc.JobControllerClient, creds credentials.DataprocCredentials) {
	log.Printf("Handling getting job")

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
