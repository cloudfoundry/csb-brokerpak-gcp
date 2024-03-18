terraform {
  required_providers {
    csbmysql = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmysql"
      version = ">= 1.2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">=3.1.0"
    }
  }
}
