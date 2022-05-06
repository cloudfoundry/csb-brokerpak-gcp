variable "db_name" { type = string }
variable "hostname" { type = string }
variable "admin_username" { type = string }
variable "admin_password" { type = string }
variable "require_ssl" { type = bool }
variable "sslcert" { type = string }
variable "sslkey" { type = string }
variable "sslrootcert" { type = string }

locals {
  port = 5432
}