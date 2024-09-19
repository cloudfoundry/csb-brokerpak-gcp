package main

import (
	"boshmanifestcreator/business/bosh"
	"boshmanifestcreator/business/capi"
	"boshmanifestcreator/business/opsmanager"
	"boshmanifestcreator/foundation/config"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	varsTemplateFile, varsFile, manifest, iaasReleasePath string
)

const (
	org   = "pivotal"
	space = "broker-cf-test"
)

func init() {
	var brokerpakPath string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}

	defaultBrokerpakPath := filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/")
	defaultIaaSReleasePath := filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/../csb-gcp-release/")

	flag.StringVar(&brokerpakPath, "brokerpak-path", defaultBrokerpakPath, "Path to the csb-brokerpak project")
	flag.StringVar(&iaasReleasePath, "iaas-release-path", defaultIaaSReleasePath, "Path to the csb-iaas-release project")
	flag.Parse()

	if brokerpakPath == "" || iaasReleasePath == "" {
		log.Fatalf("both brokerpak-path and iaas-release-path flags must be provided")
	}

	varsTemplateFile = filepath.Join(brokerpakPath, "acceptance-tests/assets/vars-template.yml")
	varsFile = filepath.Join(brokerpakPath, "acceptance-tests/assets/vars.yml")
	manifest = filepath.Join(brokerpakPath, "acceptance-tests/assets/manifest.yml")
}

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("failed to create cfg: %v", err)
	}

	// -------------------------------------------------------------------------

	if err := opsmanager.ExportEnvVariables(cfg.Toolsmiths.EnvLockMetadata); err != nil {
		log.Fatalf("failed to export environment metadata: %v", err)
	}

	// -------------------------------------------------------------------------

	ca, err := capi.New(cfg.Toolsmiths.EnvLockMetadata, org, space)
	if err != nil {
		log.Fatalf("failed to create CAPI: %v", err)
	}

	capiData, err := ca.Data()
	if err != nil {
		log.Fatalf("failed to get CAPI data: %v", err)
	}

	serviceKey, err := capi.CreateCSBServiceKey("csb-sql", "csb-sql")
	if err != nil {
		log.Fatalf("failed to create CSB service key: %v", err)
	}

	// -------------------------------------------------------------------------

	if err := bosh.CreateVarsFile(cfg, capiData, serviceKey, varsTemplateFile, varsFile); err != nil {
		log.Fatalf("failed to create vars file: %v", err)
	}

	destinationPath := fmt.Sprintf("%s/tmp-manifest.yml", os.TempDir())
	if err := bosh.CreateTempManifest(manifest, varsFile, iaasReleasePath, destinationPath); err != nil {
		log.Fatalf("failed to create temporary BOSH manifest: %v", err)
	}
}
