package capi

import (
	"fmt"
	"os"
	"os/exec"
)

func Login(envLockMetadataFilePath, org, space string) error {
	cmd := exec.Command("smith", "-l", envLockMetadataFilePath, "cf-login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to login to Cloud Foundry: %v", err)
	}

	cmd = exec.Command("cf", "target", "-o", org, "-s", space)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to target org: %v", err)
	}
	return nil
}
