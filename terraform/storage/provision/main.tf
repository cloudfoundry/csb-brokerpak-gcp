resource "google_storage_bucket" "bucket" {
  name                     = var.name
  location                 = var.region
  storage_class            = var.storage_class
  labels                   = var.labels
  public_access_prevention = var.public_access_prevention

  dynamic "custom_placement_config" {
    for_each = length(var.placement_dual_region_data_locations) == 0 ? [] : [null]
    content {
      data_locations = var.placement_dual_region_data_locations
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}
