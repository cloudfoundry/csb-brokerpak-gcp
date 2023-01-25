provider "mysql" {
  endpoint = format("%s:%d", var.private_ip, local.port)
  username = var.admin_username
  password = var.admin_password
}

provider "csbmysql" {
  database    = var.mysql_db_name
  password    = var.admin_password
  username    = var.admin_username
  port        = local.port
  host        = var.private_ip
  sslrootcert = var.sslrootcert
  sslcert     = var.sslcert
  sslkey      = var.sslkey
  skip_verify = true
}