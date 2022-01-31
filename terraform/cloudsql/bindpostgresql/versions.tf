terraform {
  required_providers {
    postgresql = {
      source  = "hashicorp/postgresql"
      version = ">=1.7.1"
    }

    random = {
      source  = "hashicorp/random"
      version = ">=3.1.0"
    }
  }
}
