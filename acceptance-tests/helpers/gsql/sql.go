// Package gsql helper functions to create and restore backups
package gsql

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/gcloud"
)

// CreateBackupBucket creates storage buckets for google sql import export procedure
func CreateBackupBucket(bucketName string) {
	gcloud.GSUtil(
		"mb",
		fmt.Sprintf("gs://%s", bucketName),
	)

}
func DeleteBucket(bucketName string) {

	gcloud.GSUtil(
		"rm",
		"-r",
		fmt.Sprintf("gs://%s/*", bucketName),
	)
	gcloud.GSUtil(
		"rb",
		fmt.Sprintf("gs://%s", bucketName),
	)
}

func getInstanceServiceAccountName(instanceID string) string {

	response := map[string]any{}

	instanceDataBytes := gcloud.GCP(
		"sql",
		"instances",
		"describe",
		instanceID,
		"--format",
		"json",
	)

	err := json.Unmarshal(instanceDataBytes, &response)
	Expect(err).ToNot(HaveOccurred())

	instanceServiceAccountName, ok := response["serviceAccountEmailAddress"].(string)
	Expect(ok).To(BeTrue())

	return instanceServiceAccountName

	// CreateBackup creates an export based backup into a target bucket
}
func CreateBackup(instanceID, targetDBName, targetBucketName string) string {

	instanceServiceAccountName := getInstanceServiceAccountName(instanceID)
	gcloud.GSUtil(
		"acl",
		"ch",
		"-u",
		fmt.Sprintf("%s:W", instanceServiceAccountName),
		fmt.Sprintf("gs://%s", targetBucketName),
	)
	dumpURI := fmt.Sprintf("gs://%s/%s.sql", targetBucketName, instanceID)
	gcloud.GCP(
		"sql",
		"export",
		"sql",
		instanceID,
		dumpURI,
		"-d",
		targetDBName,
	)

	return dumpURI

}

func RestoreBackup(dumpURI, instanceID, databaseName string) {

	instanceServiceAccountName := getInstanceServiceAccountName(instanceID)
	gcloud.GSUtil(
		"acl",
		"ch",
		"-u",
		fmt.Sprintf("%s:R", instanceServiceAccountName),
		dumpURI,
	)

	gcloud.GCP(
		"sql",
		"import",
		"sql",
		"-d",
		databaseName,
		instanceID,
		dumpURI,
		"--quiet",
	)
}
