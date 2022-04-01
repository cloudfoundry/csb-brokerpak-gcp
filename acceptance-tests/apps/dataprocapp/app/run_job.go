package app

import (
	"context"
	"dataprocapp/credentials"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	dataproc "cloud.google.com/go/dataproc/apiv1"
	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
	dataprocpb "google.golang.org/genproto/googleapis/cloud/dataproc/v1"
)

func handleRunJob(jobClient dataproc.JobControllerClient, creds credentials.DataprocCredentials) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling job run")

		jobName, ok := mux.Vars(r)["job"]
		if !ok {
			fail(w, http.StatusBadRequest, "job missing.")
			return
		}

		submitJobReq := &dataprocpb.SubmitJobRequest{
			ProjectId: creds.ProjectID,
			Region:    creds.Region,
			Job: &dataprocpb.Job{
				Reference: &dataprocpb.JobReference{
					JobId: jobName,
				},
				Placement: &dataprocpb.JobPlacement{
					ClusterName: creds.ClusterName,
				},
				TypeJob: &dataprocpb.Job_PysparkJob{
					PysparkJob: &dataprocpb.PySparkJob{
						MainPythonFileUri: fmt.Sprintf("gs://dataproc_input_for_test_do_not_delete/script.py"),
					},
				},
			},
		}

		ctx := context.Background()
		submitJobOp, err := jobClient.SubmitJobAsOperation(ctx, submitJobReq)
		if err != nil {
			fail(w, http.StatusFailedDependency, "error with request to submitting job: %v", err)
			return
		}

		submitJobResp, err := submitJobOp.Wait(ctx)
		if err != nil {
			fail(w, http.StatusFailedDependency, "error submitting job: %v", err)
			return
		}
		re := regexp.MustCompile("gs://(.+?)/(.+)")
		matches := re.FindStringSubmatch(submitJobResp.DriverOutputResourceUri)
		log.Printf("Job Response: %v", submitJobResp)
		log.Printf("Job Response DriverOutputResourceUri: %s", submitJobResp.DriverOutputResourceUri)

		if len(matches) < 3 {
			fail(w, http.StatusInternalServerError, "regex error: %s", submitJobResp.DriverOutputResourceUri)
			return
		}

		// Dataproc job outget gets saved to a GCS bucket allocated to it.
		storageClient, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(creds.Credentials))
		if err != nil {
			fail(w, http.StatusFailedDependency, "error creating storage jobClient: %v", err)
			return
		}

		obj := fmt.Sprintf("%s.000000000", matches[2])
		reader, err := storageClient.Bucket(creds.BucketName).Object(obj).NewReader(ctx)
		if err != nil {
			fail(w, http.StatusFailedDependency, "error reading job output: %v", err)
			return
		}

		defer reader.Close()

		body, err := io.ReadAll(reader)
		if err != nil {
			fail(w, http.StatusFailedDependency, "could not read output from Dataproc Job: %v", err)
			return
		}

		log.Printf("Job finished successfully: %s", body)

		w.WriteHeader(http.StatusOK)
	}
}
