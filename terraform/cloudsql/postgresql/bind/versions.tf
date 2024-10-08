terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = "~> 3"
    }

    csbpg = {
      source  = "cloudfoundry.org/cloud-service-broker/csbpg"
      version = ">=1.0.0"
    }
  }
}
