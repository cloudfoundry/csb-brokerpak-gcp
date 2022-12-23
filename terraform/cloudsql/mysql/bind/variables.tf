variable "mysql_db_name" { type = string }
variable "mysql_hostname" { type = string }
variable "private_ip" { type = string }
variable "admin_username" { type = string }
variable "admin_password" {
  sensitive = true
  type      = string
}
variable "use_tls" { type = bool }

locals {
  port = 3306
}
