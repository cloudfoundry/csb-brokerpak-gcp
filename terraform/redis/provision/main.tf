resource "google_redis_instance" "instance" {
  name               = var.instance_id
  tier               = var.service_tier
  memory_size_gb     = var.memory_size_gb
  display_name       = var.display_name
  region             = var.region
  authorized_network = local.authorized_network_id
  labels             = var.labels
  reserved_ip_range  = var.reserved_ip_range == "" ? null : var.reserved_ip_range

  timeouts {
    create = "15m"
    update = "15m"
    delete = "15m"
  }

  lifecycle {
    prevent_destroy = true
  }
}
