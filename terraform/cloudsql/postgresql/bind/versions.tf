terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = ">=3.1.0"
    }

    csbpg = {
      source  = "cloudfoundry.org/cloud-service-broker/csbpg"
      version = ">=1.0.0"
    }
  }
}
