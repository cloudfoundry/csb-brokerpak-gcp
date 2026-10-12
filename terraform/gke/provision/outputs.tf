output "cluster_name" {
  value = google_container_cluster.cluster.name
}

output "project_id" {
  value = var.project
}

output "zone" {
  value = var.zone
}

output "network" {
  value = local.network
}

output "subnetwork" {
  value = local.subnetwork == null ? "" : local.subnetwork
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}

output "persistent_disk_csi_enabled" {
  value = var.enable_persistent_disk_csi
}

output "filestore_csi_enabled" {
  value = var.enable_filestore_csi
}

output "normalized_instance_json" {
  value = jsonencode({
    version  = "v1"
    provider = "gcp"
    service  = "kubernetes"
    resource = {
      name       = google_container_cluster.cluster.name
      project_id = var.project
      location   = var.zone
      scope      = "zonal"
    }
    topology = {
      mode          = "standard"
      node_count    = 3
      machine_type  = "e2-standard-2"
      private_nodes = true
      network       = local.network
      subnetwork    = local.subnetwork == null ? "" : local.subnetwork
      storage = {
        persistent_disk_csi = var.enable_persistent_disk_csi
        filestore_csi       = var.enable_filestore_csi
      }
    }
    lifecycle = {
      ttl_hours      = var.ttl_hours
      ttl_expires_at = local.ttl_expires_at
    }
  })
}
