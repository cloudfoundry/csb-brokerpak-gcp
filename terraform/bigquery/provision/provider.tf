provider "google" {
  version     = ">=3.17.0"
  credentials = var.credentials
  project     = var.project
  region      = var.region
}

provider "google-beta" {
  version = ">=3.22.0"
  project = var.project
}
