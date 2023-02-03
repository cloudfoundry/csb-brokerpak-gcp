terraform {
  required_providers {
    mysql = {
      source  = "hashicorp/mysql"
      version = ">=1.9.0"
    }
    csbmysql = {
      source  = "cloud-service-broker/csbmysql"
      version = ">= 1.2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">=3.1.0"
    }
  }
}
