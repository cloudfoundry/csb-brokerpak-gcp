data "google_compute_network" "authorized-network" {
  name = "default"
}

locals {
  authorized_network_id           = length(var.authorized_network_id) > 0 ? var.authorized_network_id : data.google_compute_network.authorized-network.self_link
  backups_enabled                 = var.backups_retain_number > 0
  transaction_log_backups_enabled = var.backups_transaction_log_retention_days > 0
}
