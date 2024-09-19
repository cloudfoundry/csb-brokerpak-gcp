package repoassets

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Assets struct {
	VarsTemplateFilePath string
	VarsFilePath         string
	ManifestPath         string
	IaasReleasePath      string
}

func Init() (Assets, error) {
	var brokerpakPath, iaasReleasePath string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get home directory: %v", err)
	}

	defaultBrokerpakPath := filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/")
	defaultIaaSReleasePath := filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/../csb-gcp-release/")

	flag.StringVar(&brokerpakPath, "brokerpak-path", defaultBrokerpakPath, "Path to the csb-brokerpak project")
	flag.StringVar(&iaasReleasePath, "iaas-release-path", defaultIaaSReleasePath, "Path to the csb-iaas-release project")
	flag.Parse()

	if brokerpakPath == "" || iaasReleasePath == "" {
		return Assets{}, fmt.Errorf("both brokerpak-path and iaas-release-path flags must be provided")
	}

	varsTemplateFilePath, err := filepath.Abs(filepath.Join(brokerpakPath, "acceptance-tests/assets/vars-template.yml"))
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get absolute path of vars-template.yml: %v", err)
	}

	varsFilePath, err := filepath.Abs(filepath.Join(brokerpakPath, "acceptance-tests/assets/vars.yml"))
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get absolute path of vars.yml: %v", err)
	}

	manifestPath, err := filepath.Abs(filepath.Join(brokerpakPath, "acceptance-tests/assets/manifest.yml"))
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get absolute path of manifest.yml: %v", err)
	}

	iaasReleasePath, err = filepath.Abs(iaasReleasePath)
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get absolute path of iaas-release-path: %v", err)
	}

	return Assets{
		VarsTemplateFilePath: varsTemplateFilePath,
		VarsFilePath:         varsFilePath,
		ManifestPath:         manifestPath,
		IaasReleasePath:      iaasReleasePath,
	}, nil
}
