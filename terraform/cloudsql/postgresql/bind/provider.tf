provider "postgresql" {
  host        = var.hostname
  port        = local.port
  username    = var.admin_username
  password    = var.admin_password
  superuser   = false
  database    = var.db_name
  sslmode     = var.use_tls ? "verify-ca" : "disable"
  clientcert {
    cert = "${path.module}/sslcert.pem"
    key  = "${path.module}/sslkey.pem"
  }
  sslrootcert = "${path.module}/sslrootcert.pem"
}

provider "local" {}