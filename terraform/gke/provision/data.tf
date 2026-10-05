locals {
  ttl_expires_at = timeadd(timestamp(), "${var.ttl_hours}h")
  network        = trimspace(var.network) == "" ? "default" : var.network
  subnetwork     = trimspace(var.subnetwork) == "" ? null : var.subnetwork
  common_labels = merge(var.labels, {
    ttlexpiry   = lower(replace(local.ttl_expires_at, ":", "-"))
    managed-by  = "cloud-service-broker"
    environment = "sandbox"
  })

  # Service account IDs must be 6-30 chars, lowercase alphanumeric/hyphen, and
  # not end in a hyphen.
  node_service_account_id = trimsuffix(substr("csb-gke-${var.cluster_name}", 0, 30), "-")
  broker_service_account  = jsondecode(var.credentials).client_email
}
