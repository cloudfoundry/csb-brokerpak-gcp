provider "google" {
  version     = ">=3.17.0"
  credentials = var.credentials
  project     = var.project
}
