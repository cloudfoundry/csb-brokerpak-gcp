package main

import (
	"boshifier/business/bosh"
	"boshifier/business/capi"
	"boshifier/business/opsmanager"
	foundationCapi "boshifier/foundation/capi"
	"boshifier/foundation/config"
	"boshifier/foundation/flags"
	"fmt"
	"log"
	"os"
)

func main() {
	fgs, err := flags.Init()
	if err != nil {
		log.Fatalf("failed to initialize flags: %v", err)
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

	serviceKeyName := fmt.Sprintf("csb-%s", fgs.DBName)
	serviceKey, err := capi.CreateCSBServiceKey("csb-sql", serviceKeyName, map[string]string{"schema": fgs.DBName})
	if err != nil {
		log.Fatalf("failed to create CSB service key: %v", err)
	}

	// -------------------------------------------------------------------------
	boshDBBlock := bosh.CreateDBManifestBlock(serviceKey, fgs.DBSecret)

	err = bosh.CreateVarsFile(
		cfg,
		capiData,
		boshDBBlock,
		fgs.VarsTemplateFilePath,
		fgs.VarsFilePath,
	)
	if err != nil {
		log.Fatalf("failed to create vars file: %v", err)
	}

	destinationPath := fmt.Sprintf("%s/tmp-manifest.yml", os.TempDir())
	err = bosh.CreateTempManifest(
		fgs.ManifestPath,
		fgs.VarsFilePath,
		fgs.IaasReleasePath,
		destinationPath,
	)
	if err != nil {
		log.Fatalf("failed to create temporary BOSH manifest: %v", err)
	}
}
