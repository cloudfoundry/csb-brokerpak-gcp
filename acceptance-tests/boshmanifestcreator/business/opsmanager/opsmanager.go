package opsmanager

import (
	"fmt"
	"os"
	"os/exec"
)

func ExportEnvVariables(envLockMetadataFilePath string) error {
	cmd := exec.Command("smith", "om", "-l", envLockMetadataFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to source environment metadata: %v", err)
	}
	return nil
}
