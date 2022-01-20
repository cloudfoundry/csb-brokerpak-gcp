data "google_compute_network" "authorized-network" {
  name = var.authorized_network
}

locals {
  service_tiers = {
    // https://cloud.google.com/sql/pricing#2nd-gen-pricing
    1   = "db-n1-standard-1"
    2   = "db-n1-standard-2"
    4   = "db-n1-standard-4"
    8   = "db-n1-standard-8"
    16  = "db-n1-standard-16"
    32  = "db-n1-standard-32"
    64  = "db-n1-standard-64"
    0.6 = "db-f1-micro"
    1.7 = "db-g1-small"
  }
  authorized_network_id = length(var.authorized_network_id) > 0 ? var.authorized_network_id : data.google_compute_network.authorized-network.self_link
}
