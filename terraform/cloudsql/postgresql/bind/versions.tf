terraform {
  required_providers {
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = ">=1.15.0"
    }

    random = {
      source  = "hashicorp/random"
      version = ">=3.1.0"
    }
  }
}
