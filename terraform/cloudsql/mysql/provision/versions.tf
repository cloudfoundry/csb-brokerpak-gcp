terraform {
  required_providers {
    google = {
      source  = "registry.terraform.io/hashicorp/google"
      version = "~> 6"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = "~> 3"
    }
  }
}
