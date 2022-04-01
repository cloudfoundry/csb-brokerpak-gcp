variable "db_name" { type = string }
variable "hostname" { type = string }
variable "admin_username" { type = string }
variable "admin_password" { type = string }
variable "use_tls" { type = bool }

locals {
  port = 5432
}

variable "sslcert" { type = string }
variable "sslkey" { type = string }
variable "sslrootcert" { type = string }
