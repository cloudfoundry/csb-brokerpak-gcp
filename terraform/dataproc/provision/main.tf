resource "google_dataproc_cluster" "cluster" {
  name   = var.name
  region = var.region
  labels = var.labels

  cluster_config {
    master_config {
      num_instances = var.master_count
      machine_type  = var.master_machine_type
    }

    worker_config {
      num_instances = var.worker_count
      machine_type  = var.worker_machine_type
    }

    preemptible_worker_config {
      num_instances = var.preemptible_count
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}
