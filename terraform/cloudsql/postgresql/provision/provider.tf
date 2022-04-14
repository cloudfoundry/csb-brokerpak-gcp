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
  sslmode     = "verify-ca"
  sslrootcert = local_file.sslrootcert.filename
  clientcert {
    cert = local_file.sslcert.filename
    key  = local_sensitive_file.sslkey.filename
  }
}