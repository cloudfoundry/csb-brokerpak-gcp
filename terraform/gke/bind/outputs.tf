locals {
  exec_command = "gcloud"
  exec_args = [
    "container",
    "clusters",
    "get-credentials",
    var.cluster_name,
    "--project",
    var.project_id,
    "--zone",
    var.zone,
  ]
}

output "cluster_name" {
  value = var.cluster_name
}

output "project_id" {
  value = var.project_id
}

output "zone" {
  value = var.zone
}

output "ttl_expires_at" {
  value = var.ttl_expires_at
}

output "exec_command" {
  value = local.exec_command
}

output "exec_args_json" {
  value = jsonencode(local.exec_args)
}

output "normalized_binding_json" {
  value = jsonencode({
    version         = "v1"
    provider        = "gcp"
    connection_type = "cluster"
    resource = {
      name       = var.cluster_name
      project_id = var.project_id
      location   = var.zone
      scope      = "zonal"
    }
    access = {
      mode = "exec"
      exec = {
        command = local.exec_command
        args    = local.exec_args
      }
    }
    lifecycle = {
      ttl_expires_at = var.ttl_expires_at
    }
  })
}
