package main

import (
	"fmt"
	"log"
	"os"

	"boshifier/business/bosh"
	"boshifier/business/capi"
	"boshifier/foundation/config"
	"boshifier/foundation/flags"
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

	serviceKeyName := fmt.Sprintf("csb-%s", fgs.DBName)
	serviceKey, err := capi.CreateCSBServiceKey("csb-sql", serviceKeyName, map[string]string{"schema": fgs.DBName})
	if err != nil {
		log.Fatalf("failed to create CSB service key: %v", err)
	}

	// -------------------------------------------------------------------------
	boshDBBlock := bosh.CreateDBManifestBlock(serviceKey, fgs.DBSecret)

	err = bosh.CreateVarsFile(
		cfg,
		boshDBBlock,
		fgs.VarsTemplateFilePath,
		fgs.VarsFilePath,
	)
	if err != nil {
		log.Fatalf("failed to create vars file: %v", err)
	}

	// TODO destination path could be included as a flag
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