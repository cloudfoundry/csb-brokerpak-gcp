package gsql

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/gcloud"
	"encoding/json"

	"github.com/onsi/gomega"
)

func CreateBackup(instanceId string) string {

	response := map[string]any{}

	backupCreateBytes := gcloud.GCP(
		"sql",
		"backups",
		"create",
		"--instance",
		instanceId,
		"--async",
	)

	json.Unmarshal(backupCreateBytes, &response)
	operationId, ok := response["name"].(string)
	gomega.Expect(ok).To(gomega.BeTrue())

	gomega.Eventually(getOperationStatus(operationId)).Should(gomega.Equal("DONE"))

	backupContext, ok := response["backupContext"].(map[string]string)
	gomega.Expect(ok).To(gomega.BeTrue())

	backupId, ok := backupContext["backupId"]
	gomega.Expect(ok).To(gomega.BeTrue())

	return backupId

}
func RestoreBackup(sourceInstance, targetInstance, backupId string) {

	backupRestoreBytes := gcloud.GCP(
		"sql",
		"backups",
		"restore",
		backupId,
		"--restore-instance",
		targetInstance,
		"--backup-instance",
		sourceInstance,
		"--quiet",
		"--async",
		"--format",
		"json",
	)

	response := map[string]string{}

	json.Unmarshal(backupRestoreBytes, &response)
	operationId, ok := response["name"]
	gomega.Expect(ok).To(gomega.BeTrue())

	gomega.Eventually(getOperationStatus(operationId)).Should(gomega.Equal("DONE"))

}

func getOperationStatus(operationId string) string {

	statusBytes := gcloud.GCP(
		"sql",
		"operations",
		"describe",
		operationId,
		"--format",
		"json",
	)
	response := map[string]string{}

	json.Unmarshal(statusBytes, &response)
	val, ok := response["status"]
	gomega.Expect(ok).To(gomega.BeTrue())

	return val
}
