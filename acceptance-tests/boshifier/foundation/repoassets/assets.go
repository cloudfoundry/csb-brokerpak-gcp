package repoassets

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Assets struct {
	VarsTemplateFilePath   string
	VarsFilePath           string
	ManifestPath           string
	IaasReleasePath        string
	BrokerpakPath          string
	CloudServiceBrokerPath string
	TmpIaaSReleasePath     string
}

func Init() (Assets, error) {
	var brokerpakPath, iaasReleasePath, tmpIaaSReleasePath, cloudServiceBrokerPath string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get home directory: %v", err)
	}

	defaultBrokerpakPath := filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/")
	defaultIaaSReleasePath := filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/../csb-gcp-release/")
	defaultCloudServiceBrokerPath := filepath.Join(homeDir, "workspace/csb/cloud-service-broker/")
	defaultTmpIaaSReleasePath := "/tmp/csb-gcp-release"

	flag.StringVar(&brokerpakPath, "brokerpak-path", defaultBrokerpakPath, "Path to the csb-brokerpak project")
	flag.StringVar(&cloudServiceBrokerPath, "cloud-service-broker-path", defaultCloudServiceBrokerPath, "Path to the cloud-service-broker project")
	flag.StringVar(&iaasReleasePath, "iaas-release-path", defaultIaaSReleasePath, "Path to the csb-iaas-release project")
	flag.StringVar(&tmpIaaSReleasePath, "tmp-release-path", defaultTmpIaaSReleasePath, "Path to the destination release project where we will copy our modified release")

	flag.Parse()

	if brokerpakPath == "" || iaasReleasePath == "" || tmpIaaSReleasePath == "" || cloudServiceBrokerPath == "" {
		return Assets{}, fmt.Errorf("both brokerpak-path, iaas-release-path, tmp-release-path, and cloud-service-broker-path flags must be provided")
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

	cloudServiceBrokerPath, err = filepath.Abs(cloudServiceBrokerPath)
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get absolute path of cloud-service-broker-path: %v", err)
	}

	tmpIaaSReleasePath, err = filepath.Abs(tmpIaaSReleasePath)
	if err != nil {
		return Assets{}, fmt.Errorf("failed to get absolute path of tmp-release-path: %v", err)
	}

	return Assets{
		VarsTemplateFilePath:   varsTemplateFilePath,
		VarsFilePath:           varsFilePath,
		ManifestPath:           manifestPath,
		IaasReleasePath:        iaasReleasePath,
		BrokerpakPath:          brokerpakPath,
		CloudServiceBrokerPath: cloudServiceBrokerPath,
		TmpIaaSReleasePath:     tmpIaaSReleasePath,
	}, nil
}
