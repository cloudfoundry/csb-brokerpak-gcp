package main

import (
	"log"

	"boshifier/business/bosh"
	"boshifier/foundation/config"
	"boshifier/foundation/flags"
)

func main() {
	fgs, err := flags.Init()
	if err != nil {
		log.Fatalf("failed to initialize flags: %v", err)
	}

	if err := config.Check(); err != nil {
		log.Fatalf("failed to check cfg: %v", err)
	}

	// -------------------------------------------------------------------------

	if err = bosh.Deploy(fgs.BoshDeploymentName, fgs.ManifestPath, fgs.VarsFilePath, fgs.IaasReleasePath); err != nil {
		log.Fatalf("failed to deploy: %v", err)
	}
}
