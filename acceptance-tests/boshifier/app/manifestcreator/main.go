package main

import (
	"boshifier/business/bosh"
	"boshifier/business/capi"
	"boshifier/business/opsmanager"
	foundationCapi "boshifier/foundation/capi"
	"boshifier/foundation/config"
	"boshifier/foundation/repoassets"
	"fmt"
	"log"
	"os"
)

func main() {
	assets, err := repoassets.Init()
	if err != nil {
		log.Fatalf("failed to initialize assets: %v", err)
	}

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("failed to create cfg: %v", err)
	}

	// -------------------------------------------------------------------------

	if err := opsmanager.ExportEnvVariables(cfg.Toolsmiths.EnvLockMetadata); err != nil {
		log.Fatalf("failed to export environment metadata: %v", err)
	}

	// -------------------------------------------------------------------------

	ca, err := capi.New(cfg.Toolsmiths.EnvLockMetadata, foundationCapi.Organization, foundationCapi.Space)
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

	err = bosh.CreateVarsFile(
		cfg,
		capiData,
		serviceKey,
		assets.VarsTemplateFilePath,
		assets.VarsFilePath,
	)
	if err != nil {
		log.Fatalf("failed to create vars file: %v", err)
	}

	destinationPath := fmt.Sprintf("%s/tmp-manifest.yml", os.TempDir())
	err = bosh.CreateTempManifest(
		assets.ManifestPath,
		assets.VarsFilePath,
		assets.IaasReleasePath,
		destinationPath,
	)
	if err != nil {
		log.Fatalf("failed to create temporary BOSH manifest: %v", err)
	}
}
