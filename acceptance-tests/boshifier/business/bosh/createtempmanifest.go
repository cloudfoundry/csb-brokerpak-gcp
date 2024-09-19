package bosh

import (
	"fmt"
	"os"
	"os/exec"
)

func CreateTempManifest(manifestPath, varsFilePath, iaasReleasePath, destinationPath string) error {
	cmd := exec.Command(
		"bosh",
		"int",
		manifestPath,
		"-l",
		varsFilePath,
		"-v",
		fmt.Sprintf("release_repo_path=%s", iaasReleasePath),
	)

	tmpManifestFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary manifest file: %v", err)
	}
	defer tmpManifestFile.Close()

	cmd.Stdout = tmpManifestFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create temporary BOSH manifest: %v", err)
	}
	return nil
}
