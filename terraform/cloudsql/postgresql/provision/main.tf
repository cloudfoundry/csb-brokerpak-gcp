resource "google_sql_database_instance" "instance" {
  name             = var.instance_name
  database_version = var.database_version
  region           = var.region

  settings {
    tier        = var.tier
    disk_size   = var.storage_gb
    disk_autoresize_limit = var.storage_autoresize_limit
    user_labels = var.labels

    availability_type = local.availability_type

    ip_configuration {
      ipv4_enabled    = var.public_ip
      private_network = local.authorized_network_id
      require_ssl     = var.require_ssl

      dynamic "authorized_networks" {
        for_each = var.authorized_networks_cidrs
        iterator = networks

        content {
          value = networks.value
        }
      }
    }

    database_flags {
      name  = "password_encryption"
      value = "scram-sha-256"
    }

    location_preference {
      zone           = local.primary_zone
      secondary_zone = local.secondary_zone
    }

    backup_configuration {
      enabled                        = var.backups_retain_number != 0
      start_time                     = var.backups_start_time
      location                       = var.backups_location
      point_in_time_recovery_enabled = var.backups_retain_number != 0 && var.backups_point_in_time_log_retain_days != 0
      transaction_log_retention_days = var.backups_point_in_time_log_retain_days
      backup_retention_settings {
        retained_backups = var.backups_retain_number
      }
    }
  }

  deletion_protection = false

  lifecycle {
    prevent_destroy = true
    ignore_changes = [ 
      // disk_size and disk_autoresize properties do not play along well together.
      // The value of disk_size will attempt override the resizing from disk_autoresize when that feature is enabled. 
      // This will cause an attempt to recreate the database instance.
      // Ideally we would like to be able to control the lifecycle.ignore_changes depending on the value of `disk_autoresize`
      // but dynamic lifecycle.ignore_changes variables are not alloweed. See https://github.com/opentofu/opentofu/issues/1432 
      settings[0].disk_size,
    ]
  }

  timeouts {
    create = "60m"
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
  name            = random_string.username.result
  instance        = google_sql_database_instance.instance.name
  password        = random_password.password.result
  deletion_policy = "ABANDON"
}

resource "google_sql_ssl_cert" "client_cert" {
  common_name = random_string.username.result
  instance    = google_sql_database_instance.instance.name
}
