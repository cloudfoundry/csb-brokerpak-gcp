data "google_compute_network" "authorized-network" {
  count = length(var.authorized_network_id) > 0 ? 0 : 1
  name  = var.authorized_network
}

locals {
  authorized_network_id = length(var.authorized_network_id) > 0 ? var.authorized_network_id : data.google_compute_network.authorized-network[0].self_link
  availability_type     = var.highly_available ? "REGIONAL" : "ZONAL"
  primary_zone          = var.location_preference_zone == "" ? "" : join("-", [var.region, var.location_preference_zone])
  secondary_zone        = var.location_preference_secondary_zone == "" ? "" : join("-", [var.region, var.location_preference_secondary_zone])
}
