provider "google" {
  credentials = var.credentials
  project     = var.project
  region      = var.region
}
