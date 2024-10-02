package main

import (
	"log"
	"os"
	"os/exec"

	"boshifier/business/vendir"
	"boshifier/foundation/flags"
)

func main() {
	assets, err := flags.Init()
	if err != nil {
		log.Fatalf("failed to initialize assets: %v", err)
	}

	// -------------------------------------------------------------------------

	if err := removeTmpIaaSReleasePath(assets.TmpIaaSReleasePath); err != nil {
		log.Fatalf("failed to remove existing tmp iaas release path: %v", err)
	}

	if err := copyIaaSReleasePath(assets.IaasReleasePath, assets.TmpIaaSReleasePath); err != nil {
		log.Fatalf("failed to copy iaas-release-path to %s: %v", assets.TmpIaaSReleasePath, err)
	}

	// -------------------------------------------------------------------------

	csbPkgName, err := vendir.FindPackagePathByURL("cloud-service-broker", assets.TmpIaaSReleasePath)
	if err != nil {
		log.Fatalf("failed to find cloud-service-broker package path: %v", err)
	}

	err = vendir.Sync(assets.TmpIaaSReleasePath, assets.BrokerpakPath, csbPkgName, assets.CloudServiceBrokerPath)
	if err != nil {
		log.Fatalf("failed to sync vendir: %v", err)
	}

	if err := vendir.GoModVendoringPackages(assets.TmpIaaSReleasePath); err != nil {
		log.Fatalf("failed to vendor packages: %v", err)
	}
}

func removeTmpIaaSReleasePath(path string) error {
	if _, err := os.Stat(path); err == nil {
		return os.RemoveAll(path)
	}
	return nil
}

func copyIaaSReleasePath(src, dst string) error {
	cmd := exec.Command("cp", "-r", src, dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}