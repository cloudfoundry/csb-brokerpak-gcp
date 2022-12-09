resource "google_sql_database_instance" "instance" {
  name                = var.instance_name
  database_version    = var.database_version
  region              = var.region
  deletion_protection = var.deletion_protection

  settings {
    tier        = var.tier
    disk_size   = var.storage_gb
    user_labels = var.labels

    ip_configuration {
      ipv4_enabled    = false
      private_network = local.authorized_network_id
      #require_ssl = var.use_tls
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
  length           = 16
  special          = true
  override_special = "_@"
}

resource "google_sql_user" "admin_user" {
  name     = random_string.username.result
  instance = google_sql_database_instance.instance.name
  password = random_password.password.result
}
