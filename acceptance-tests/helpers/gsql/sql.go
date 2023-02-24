// Package gsql helper functions to create and restore backups
package gsql

import (
	"crypto/sha256"
	"csbbrokerpakgcp/acceptance-tests/helpers/gcloud"
	"encoding/json"
	"fmt"
	"os"

	. "github.com/onsi/gomega"
)

// CreateBackupBucket creates storage buckets for google sql import export procedure
func CreateBackupBucket(bucketName string) {
	gcloud.GSUtil(
		"mb",
		fmt.Sprintf("gs://%s", bucketName),
	)
}

// DeleteBucket deletes a bucket, along with all its contents
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
}

// CreateBackup creates an export based backup into a target bucket
func CreateBackup(instanceID, dbName, bucketName string) string {
	enableBucketWrite(getInstanceServiceAccountName(instanceID), bucketName)
	dumpURI := fmt.Sprintf("gs://%s/%s.sql", bucketName, instanceID)
	gcloud.GCP(
		"sql",
		"export",
		"sql",
		instanceID,
		dumpURI,
		"-d",
		dbName,
	)

	return dumpURI
}

func enableBucketWrite(serviceAccountEmail, bucketName string) {
	gcloud.GSUtil(
		"acl",
		"ch",
		"-u",
		fmt.Sprintf("%s:W", serviceAccountEmail),
		fmt.Sprintf("gs://%s", bucketName),
	)
}

func enableFileRead(serviceAccountEmail, fileURI string) {
	gcloud.GSUtil(
		"acl",
		"ch",
		"-u",
		fmt.Sprintf("%s:R", serviceAccountEmail),
		fileURI,
	)
}

// PerformAdminSQL executes SQL against a database in a CloudSQL instance, with platform admin privileges
//
//	further info: https://cloud.google.com/sql/docs/mysql/import-export/import-export-sql#gcloud_1
func PerformAdminSQL(queryString, instanceName, dbName, bucketName string) {
	PerformSQL(queryString, instanceName, dbName, bucketName, "")
}

// PerformSQL executes SQL against a database in a CloudSQL instance
//
//	further info: https://cloud.google.com/sql/docs/mysql/import-export/import-export-sql#gcloud_1
func PerformSQL(queryString, instanceName, dbName, bucketName, userName string) {
	fileName := fmt.Sprintf("%s-%x.sql", instanceName, sha256.Sum256([]byte(queryString)))
	fileURI := fmt.Sprintf("gs://%s/%s", bucketName, fileName)

	serviceAccountName := getInstanceServiceAccountName(instanceName)
	enableBucketWrite(serviceAccountName, bucketName)

	UploadTextFile(fileURI, queryString)

	args := []string{
		"sql",
		"import",
		"sql",
		instanceName,
		fileURI,
		"-d",
		dbName,
		"--quiet",
	}

	enableFileRead(serviceAccountName, fileURI)

	if userName != "" {
		args = append(args, "--user", userName)
	}

	gcloud.GCP(args...)
}

// UploadTextFile uploads a text file to a given GCS bucket
func UploadTextFile(fileURL, contents string) {
	tempFile, err := os.CreateTemp("", "bucket-file")
	Expect(err).NotTo(HaveOccurred())
	defer os.Remove(tempFile.Name())
	_, err = tempFile.WriteString(contents)
	Expect(err).NotTo(HaveOccurred())
	gcloud.GSUtil("cp", tempFile.Name(), fileURL)
}

// RestoreBackupWithUser restores a CloudSQL database backup from a SQL file in a bucket
func RestoreBackupWithUser(dumpURI, instanceID, databaseName, username string) {
	instanceServiceAccountName := getInstanceServiceAccountName(instanceID)
	enableFileRead(instanceServiceAccountName, dumpURI)

	gcloud.GCP(
		"sql",
		"import",
		"sql",
		"-d",
		databaseName,
		instanceID,
		dumpURI,
		"--quiet",
		"--user",
		username,
	)
}

// RestoreBackup restores a CloudSQL database backup from a SQL file in a bucket
func RestoreBackup(dumpURI, instanceID, databaseName string) {
	instanceServiceAccountName := getInstanceServiceAccountName(instanceID)
	enableFileRead(instanceServiceAccountName, dumpURI)

	gcloud.GCP(
		"sql",
		"import",
		"sql",
		instanceID,
		dumpURI,
		"-d",
		databaseName,
		"--quiet",
	)
}
