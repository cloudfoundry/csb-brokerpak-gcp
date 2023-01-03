provider "csbpg" {
  host            = var.private_ip
  port            = local.port
  username        = var.admin_username
  password        = var.admin_password
  database        = var.db_name
  data_owner_role = "binding_user_group"
  sslmode         = "verify-ca"
  sslrootcert     = var.sslrootcert
  clientcert {
    cert = var.sslcert
    key  = var.sslkey
  }
}