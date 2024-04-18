package app

import (
	"context"
	"dataprocapp/credentials"
	"log"
	"net/http"

	dataproc "cloud.google.com/go/dataproc/v2/apiv1"
	"cloud.google.com/go/dataproc/v2/apiv1/dataprocpb"
)

func handleDeleteJob(w http.ResponseWriter, jobName string, jobClient *dataproc.JobControllerClient, creds credentials.DataprocCredentials) {
	log.Printf("Handling deletion of job")

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
