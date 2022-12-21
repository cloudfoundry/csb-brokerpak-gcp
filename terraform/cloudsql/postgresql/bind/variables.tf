variable "db_name" { type = string }
variable "hostname" { type = string }
variable "admin_username" { type = string }
variable "admin_password" {
  sensitive = true
  type = string
}
variable "private_ip" { type = string }
variable "require_ssl" { type = bool }
variable "sslcert" { type = string }
variable "sslkey" {
  sensitive = true
  type = string
}
variable "sslrootcert" { type = string }

locals {
  port = 5432
}