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
		log.Fatalf("failed to check config: %v", err)
	}

	// -------------------------------------------------------------------------

	if err = bosh.Delete(fgs.BoshDeploymentName); err != nil {
		log.Fatalf("failed to delete BOSH deployment: %v", err)
	}
}
