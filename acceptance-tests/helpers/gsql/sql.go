package gsql

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/gomega"

	"csbbrokerpakgcp/acceptance-tests/helpers/gcloud"
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
		"--format",
		"json",
	)

	err := json.Unmarshal(backupCreateBytes, &response)
	Expect(err).NotTo(HaveOccurred())
	operationId, ok := response["name"].(string)
	Expect(ok).To(BeTrue())
	Eventually(func() string { return getOperationStatus(operationId) }).
		WithTimeout(5 * time.Minute).
		WithPolling(10 * time.Second).
		Should(Equal("DONE"))

	Expect(response["backupContext"]).To(BeAssignableToTypeOf(map[string]any{}))
	backupContext := response["backupContext"].(map[string]any)

	Expect(backupContext["backupId"]).To(BeAssignableToTypeOf("string"))
	backupId := backupContext["backupId"].(string)

	return backupId

}

func GetPrimaryAddress(instance string) (string, error) {
	instanceDescriptionBytes := gcloud.GCP(
		"sql",
		"instances",
		"describe",
		instance,
		"--format",
		"json",
	)
	description := struct {
		IpAddresses []struct {
			IpAddress string `json:"ipAddress"`
			Type      string `json:"type"`
		} `json:"ipAddresses"`
	}{}
	err := json.Unmarshal(instanceDescriptionBytes, &description)
	if err != nil {
		return "", err
	}

	for _, addr := range description.IpAddresses {
		if addr.Type == "PRIMARY" {
			return addr.IpAddress, nil
		}
	}

	return "", fmt.Errorf("primary address not present in %#v", description)
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

	response := map[string]any{}

	err := json.Unmarshal(backupRestoreBytes, &response)
	Expect(err).NotTo(HaveOccurred())
	Expect(response["name"]).To(BeAssignableToTypeOf("string"))
	operationId := response["name"].(string)

	Eventually(func() string { return getOperationStatus(operationId) }).
		WithPolling(30 * time.Second).
		WithTimeout(30 * time.Minute).
		Should(Equal("DONE"))

}
func DeleteBackup(instanceId, backupId string) {
	gcloud.GCP(
		"sql",
		"backups",
		"delete",
		backupId,
		"--instance",
		instanceId,
		"--async",
		"--quiet",
	)
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
	response := map[string]any{}

	err := json.Unmarshal(statusBytes, &response)
	Expect(err).NotTo(HaveOccurred())
	val, ok := response["status"]
	Expect(ok).To(BeTrue())
	Expect(val).To(BeAssignableToTypeOf(""))
	return val.(string)
}
