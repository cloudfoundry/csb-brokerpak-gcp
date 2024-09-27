package main

import (
	"boshifier/business/bosh"
	"boshifier/business/opsmanager"
	"boshifier/foundation/config"
	"boshifier/foundation/flags"
	"log"
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

	if err = bosh.Delete(fgs.BoshDeploymentName); err != nil {
		log.Fatalf("failed to delete BOSH deployment: %v", err)
	}
}
