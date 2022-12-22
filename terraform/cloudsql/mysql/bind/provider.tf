provider "mysql" {
  endpoint = format("%s:%d", var.private_ip, local.port)
  username = var.admin_username
  password = var.admin_password
}
