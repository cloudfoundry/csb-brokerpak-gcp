provider "google" {
  credentials = var.credentials
  project     = var.project
}

provider "postgresql" {
  host        = google_sql_database_instance.instance.first_ip_address
  port        = 5432
  username    = google_sql_user.admin_user.name
  password    = google_sql_user.admin_user.password
  superuser   = false
  database    = google_sql_database.database.name
  sslmode     = var.use_tls ? "verify-ca" : "disable"
  clientcert {
    cert = "${path.module}/sslcert.pem"
    key  = "${path.module}/sslkey.pem"
  }
  sslrootcert = "${path.module}/sslrootcert.pem"
}

provider "local" {}