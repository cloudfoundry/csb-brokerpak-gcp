provider "google" {
  credentials = var.credentials
  project     = var.project
  zone        = var.zone
}
