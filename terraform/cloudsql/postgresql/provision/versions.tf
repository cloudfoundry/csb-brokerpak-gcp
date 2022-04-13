terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.8.0"
    }

    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = ">=1.15.0"
    }
  }
}
