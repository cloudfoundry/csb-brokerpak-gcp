data "google_compute_network" "authorized-network" {
  name = "default"
}

locals {
  authorized_network_id           = length(var.authorized_network_id) > 0 ? var.authorized_network_id : data.google_compute_network.authorized-network.self_link
  backups_enabled                 = var.backups_retain_number > 0
  transaction_log_backups_enabled = var.backups_transaction_log_retention_days > 0
  availability_type = var.highly_available ? "REGIONAL" : "ZONAL"
  primary_zone = var.location_preference_zone == "" ? "" : join("-", [var.region, var.location_preference_zone])
  secondary_zone = var.location_preference_secondary_zone == "" ? "" : join("-", [var.region, var.location_preference_secondary_zone])
}
