provider "postgresql" {
  host        = var.hostname
  port        = local.port
  username    = var.admin_username
  password    = var.admin_password
  superuser   = false
  database    = var.db_name
  sslmode     = "verify-ca"
  sslrootcert = local_file.sslrootcert.filename
  clientcert {
    cert = local_file.sslcert.filename
    key  = local_sensitive_file.sslkey.filename
  }
}
