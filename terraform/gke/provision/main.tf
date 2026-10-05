resource "google_project_service" "compute" {
  project            = var.project
  service            = "compute.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "container" {
  project            = var.project
  service            = "container.googleapis.com"
  disable_on_destroy = false
}

resource "google_container_cluster" "cluster" {
  name                     = var.cluster_name
  project                  = var.project
  location                 = var.zone
  network                  = local.network
  subnetwork               = local.subnetwork
  deletion_protection      = false
  remove_default_node_pool = true
  initial_node_count       = 1
  enable_shielded_nodes    = true
  networking_mode          = "VPC_NATIVE"
  datapath_provider        = "ADVANCED_DATAPATH"
  logging_service          = "logging.googleapis.com/kubernetes"
  monitoring_service       = "monitoring.googleapis.com/kubernetes"
  resource_labels          = local.common_labels

  ip_allocation_policy {}

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
  }

  workload_identity_config {
    workload_pool = "${var.project}.svc.id.goog"
  }

  release_channel {
    channel = "REGULAR"
  }

  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  addons_config {
    http_load_balancing {
      disabled = true
    }

    horizontal_pod_autoscaling {
      disabled = true
    }

  }

  depends_on = [
    google_project_service.compute,
    google_project_service.container,
  ]
}

resource "google_container_node_pool" "primary" {
  name       = "sandbox-pool"
  project    = var.project
  location   = var.zone
  cluster    = google_container_cluster.cluster.name
  node_count = 3

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type = "e2-standard-2"
    labels       = local.common_labels
    oauth_scopes = ["https://www.googleapis.com/auth/cloud-platform"]

    metadata = {
      disable-legacy-endpoints = "true"
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }
  }
}
