resource "google_sql_database_instance" "instance" {
  name                = var.instance_name
  database_version    = var.mysql_version
  region              = var.region
  deletion_protection = var.deletion_protection

  settings {
    tier                  = var.tier
    user_labels           = var.labels
    disk_size             = var.disk_autoresize ? null : var.storage_gb
    disk_autoresize       = var.disk_autoresize
    disk_autoresize_limit = var.disk_autoresize ? var.disk_autoresize_limit : 0
    availability_type     = local.availability_type


    ip_configuration {
      ipv4_enabled    = var.public_ip
      private_network = local.authorized_network_id

      ssl_mode = var.allow_insecure_connections ? "ALLOW_UNENCRYPTED_AND_ENCRYPTED" : "TRUSTED_CLIENT_CERTIFICATE_REQUIRED"

      dynamic "authorized_networks" {
        for_each = var.authorized_networks_cidrs
        iterator = networks

        content {
          value = networks.value
        }
      }
    }

    location_preference {
      zone           = local.primary_zone
      secondary_zone = local.secondary_zone
    }

    backup_configuration {
      enabled    = local.backups_enabled
      start_time = var.backups_start_time
      backup_retention_settings {
        retained_backups = var.backups_retain_number
        retention_unit   = "COUNT"
      }
      binary_log_enabled             = local.transaction_log_backups_enabled
      transaction_log_retention_days = var.backups_transaction_log_retention_days
    }

    dynamic "maintenance_window" {
    for_each = var.maintenance_day != null ? [null] : []
      content {
        day  = var.maintenance_day
        hour = var.maintenance_hour
      }
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_sql_database" "database" {
  name     = var.db_name
  instance = google_sql_database_instance.instance.name

  lifecycle {
    prevent_destroy = true
  }
}

resource "random_string" "username" {
  length  = 16
  special = false
}

resource "random_password" "password" {
  length           = 64
  special          = true
  override_special = "_@"
}

resource "google_sql_user" "admin_user" {
  name     = random_string.username.result
  instance = google_sql_database_instance.instance.name
  password = random_password.password.result
}

resource "google_sql_ssl_cert" "client_cert" {
  common_name = random_string.username.result
  instance    = google_sql_database_instance.instance.name
}
