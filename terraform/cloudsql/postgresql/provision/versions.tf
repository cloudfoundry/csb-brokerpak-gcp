terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.8.0"
    }

    random = {
      source  = "hashicorp/random"
      version = ">=3.1.0"
    }

    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = ">=1.15.0"
    }

    local = {
      source  = "hashicorp/local"
      version = ">=2.2.2"
    }
  }
}
