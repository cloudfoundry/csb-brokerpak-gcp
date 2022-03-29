provider "mysql" {
  endpoint = format("%s:%d", var.mysql_hostname, local.port)
  username = var.admin_username
  password = var.admin_password
}
