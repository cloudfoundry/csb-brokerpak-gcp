terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = ">=3.1.0"
    }

    csbpg = {
      source  = "cloud-service-broker/csbpg"
      version = "1.0.0"
    }
  }
}
