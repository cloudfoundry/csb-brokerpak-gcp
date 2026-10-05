locals {
  ttl_expires_at = timeadd(timestamp(), "${var.ttl_hours}h")
  network        = trimspace(var.network) == "" ? "default" : var.network
  subnetwork     = trimspace(var.subnetwork) == "" ? null : var.subnetwork
  common_labels = merge(var.labels, {
    ttlexpiry   = lower(replace(local.ttl_expires_at, ":", "-"))
    managed-by  = "cloud-service-broker"
    environment = "sandbox"
  })
}
