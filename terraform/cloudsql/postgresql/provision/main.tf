resource "google_sql_database_instance" "instance" {
  name             = var.instance_name
  database_version = var.database_version
  region           = var.region

  settings {
    tier        = local.service_tiers[var.cores]
    disk_size   = var.storage_gb
    user_labels = var.labels

    ip_configuration {
      ipv4_enabled    = var.public_ip
      private_network = local.authorized_network_id
      #require_ssl = var.use_tls

      dynamic "authorized_networks" {
        for_each = var.authorized_networks_cidrs
        iterator = networks

        content {
          value = networks.value
        }
      }
    }

    database_flags {
      name = "password_encryption"
      value = "scram-sha-256"
    }

    backup_configuration {
      enabled = var.backups_retain_number != 0
      start_time = var.backups_start_time
      location = var.backups_location
      point_in_time_recovery_enabled = var.backups_retain_number != 0 && var.backups_point_in_time_log_retain_days != 0
      transaction_log_retention_days = var.backups_point_in_time_log_retain_days
      backup_retention_settings {
        retained_backups = var.backups_retain_number
      }
    }
  }

  deletion_protection = false
}

resource "google_sql_database" "database" {
  name     = var.db_name
  instance = google_sql_database_instance.instance.name
}

resource "random_string" "username" {
  length  = 16
  special = false
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "_@"
}

resource "google_sql_user" "admin_user" {
  name     = random_string.username.result
  instance = google_sql_database_instance.instance.name
  password = random_password.password.result
  deletion_policy="ABANDON"
}

resource "random_string" "createrole_username" {
  length  = 16
  special = false
}
resource "random_password" "createrole_password" {
  length           = 16
  special          = true
  override_special = "_@"
}

resource "postgresql_role" "createrole_user" {
  depends_on  = [google_sql_user.admin_user]
  name                = random_string.createrole_username.result
  password            = random_password.createrole_password.result
  login               = true
  create_role         = true
}

resource "postgresql_grant" "db_access" {
  depends_on  = [postgresql_role.createrole_user]
  database    = var.db_name
  role        = postgresql_role.createrole_user.name
  object_type = "database"
  privileges  = ["ALL"]
}

resource "postgresql_grant" "table_access" {
  depends_on  = [postgresql_role.createrole_user]
  database    = var.db_name
  role        = postgresql_role.createrole_user.name
  schema      = "public"
  object_type = "table"
  privileges  = ["ALL"]
}
