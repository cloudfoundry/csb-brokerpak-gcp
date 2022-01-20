data "google_compute_network" "authorized-network" {
  name = var.authorized_network
}

locals {
  authorized_network_id = length(var.authorized_network_id) > 0 ? var.authorized_network_id : data.google_compute_network.authorized-network.self_link
}
