terraform {
  required_providers {
    google = {
      source  = "registry.terraform.io/hashicorp/google"
      version = ">=4.8.0"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">=3.4.3"
    }
  }
}
