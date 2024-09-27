package bosh

import (
	"fmt"
	"os"
	"os/exec"
)

func Delete(deploymentName string) error {
	cmd := exec.Command("bosh", "-n", "delete-deployment", "-d", deploymentName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start delete-deployment %s command: %v", deploymentName, err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for delete-deployment %s command to finish: %v", deploymentName, err)
	}

	return nil
}
