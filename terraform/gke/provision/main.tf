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

# Dedicated, least-privilege node service account. This project's org policy
# prevents the implicit GCE default service account from being created, so
# GKE node pools must be given an explicit service account rather than
# relying on that default.
resource "google_service_account" "nodes" {
  project      = var.project
  account_id   = local.node_service_account_id
  display_name = "CSB GKE sandbox node pool — ${var.cluster_name}"
}

resource "google_project_iam_member" "nodes_log_writer" {
  project = var.project
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.nodes.email}"
}

resource "google_project_iam_member" "nodes_metric_writer" {
  project = var.project
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.nodes.email}"
}

resource "google_project_iam_member" "nodes_monitoring_viewer" {
  project = var.project
  role    = "roles/monitoring.viewer"
  member  = "serviceAccount:${google_service_account.nodes.email}"
}

resource "google_project_iam_member" "nodes_artifact_reader" {
  project = var.project
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.nodes.email}"
}

resource "google_service_account_iam_member" "broker_can_use_nodes" {
  service_account_id = google_service_account.nodes.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${local.broker_service_account}"
}

resource "google_container_cluster" "cluster" {
  name                  = var.cluster_name
  project               = var.project
  location              = var.zone
  network               = local.network
  subnetwork            = local.subnetwork
  deletion_protection   = false
  initial_node_count    = 3
  enable_shielded_nodes = true
  networking_mode       = "VPC_NATIVE"
  datapath_provider     = "ADVANCED_DATAPATH"
  resource_labels       = local.common_labels

  ip_allocation_policy {}

  node_config {
    machine_type    = "e2-standard-2"
    labels          = local.common_labels
    service_account = google_service_account.nodes.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

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

  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]
  }

  addons_config {
    http_load_balancing {
      disabled = true
    }

    horizontal_pod_autoscaling {
      disabled = true
    }

    gce_persistent_disk_csi_driver_config {
      enabled = var.enable_persistent_disk_csi
    }

    gcp_filestore_csi_driver_config {
      enabled = var.enable_filestore_csi
    }
  }

  depends_on = [
    google_project_service.compute,
    google_project_service.container,
    google_project_iam_member.nodes_log_writer,
    google_project_iam_member.nodes_metric_writer,
    google_project_iam_member.nodes_monitoring_viewer,
    google_project_iam_member.nodes_artifact_reader,
    google_service_account_iam_member.broker_can_use_nodes,
  ]
}
