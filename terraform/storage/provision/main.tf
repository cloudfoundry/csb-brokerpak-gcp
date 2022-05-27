resource "google_storage_bucket" "bucket" {
  name          = var.name
  location      = var.region
  storage_class = var.storage_class
  labels        = var.labels

  lifecycle {
    prevent_destroy = true
  }
}
