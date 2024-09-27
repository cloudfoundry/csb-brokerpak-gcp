package vendir

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func FindPackagePathByURL(targetPackage, releasePath string) (string, error) {

	vendirSpecAbsPath, err := filepath.Abs(filepath.Join(releasePath, Filename))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path of %s: %v", Filename, err)
	}

	file, err := os.ReadFile(vendirSpecAbsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read vendir spec: %v", err)
	}

	var c Config
	if err := yaml.Unmarshal(file, &c); err != nil {
		return "", fmt.Errorf("failed to unmarshal vendir spec: %v", err)
	}

	for _, directory := range c.Directories {
		for _, content := range directory.Contents {
			if strings.Contains(content.Git.URL, targetPackage) {
				return content.Path, nil
			}
		}
	}

	return "", fmt.Errorf("package %s not found in vendir spec", targetPackage)
}

func Sync(tmpIaaSReleasePath, brokerpakPath, cloudServiceBrokerPackageName, cloudServiceBrokerPath string) error {

	cmd := exec.Command(
		"vendir",
		"sync",
		"--directory",
		fmt.Sprintf("src/csb-brokerpak-gcp=%s", brokerpakPath),
		"--directory",
		fmt.Sprintf("src/%s=%s", cloudServiceBrokerPackageName, cloudServiceBrokerPath),
	)
	cmd.Dir = tmpIaaSReleasePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute vendir sync: %v", err)
	}

	return nil
}

func GoModVendoringPackages(releasePath string) error {
	vendirSpecAbsPath, err := filepath.Abs(filepath.Join(releasePath, Filename))
	if err != nil {
		return fmt.Errorf("failed to get absolute path of %s: %v", Filename, err)
	}

	file, err := os.ReadFile(vendirSpecAbsPath)
	if err != nil {
		return fmt.Errorf("failed to read vendir spec: %v", err)
	}

	var c Config
	if err := yaml.Unmarshal(file, &c); err != nil {
		return fmt.Errorf("failed to unmarshal vendir spec: %v", err)
	}

	for _, directory := range c.Directories {
		sourcePath := filepath.Join(releasePath, directory.Path)
		for _, content := range directory.Contents {
			if err := vendorPackage(content, sourcePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func vendorPackage(content ContentDirectory, sourcePath string) error {
	packagePath, err := filepath.Abs(filepath.Join(sourcePath, content.Path))
	if err != nil {
		return fmt.Errorf("failed to get absolute path of package %s: %v", content.Path, err)
	}

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return fmt.Errorf("package %s not found in release path %s", content.Path, packagePath)
	}

	if err := goVendorCommand(packagePath); err != nil {
		return fmt.Errorf("failed to run vendir sync: %v", err)
	}

	return nil
}

func goVendorCommand(packagePath string) error {
	cmd := exec.Command("go", "mod", "vendor")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = packagePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go mod vendor in %s: %v", err, packagePath)
	}

	return nil

}
