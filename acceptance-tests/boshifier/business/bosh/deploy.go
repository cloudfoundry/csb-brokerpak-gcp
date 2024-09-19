package bosh

import (
	"fmt"
	"os"
	"os/exec"
)

func Deploy(deploymentName, manifestPath, varsFilePath, releaseRepoPath string) error {
	cmd := exec.Command(
		"bosh",
		"-d", deploymentName,
		"deploy", manifestPath,
		"-l", varsFilePath,
		"-v", fmt.Sprintf("name=%s", deploymentName),
		"-v", fmt.Sprintf("release_repo_path=%s", releaseRepoPath),
		"--no-redact",
		"-n",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start deploy command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for deploy command to finish: %v", err)
	}

	return nil
}
