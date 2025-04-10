variable "mysql_db_name" { type = string }
variable "mysql_hostname" { type = string }
variable "private_ip" { type = string }
variable "admin_username" { type = string }
variable "admin_password" {
  sensitive = true
  type      = string
}

variable "sslrootcert" { type = string }
variable "sslcert" { type = string }
variable "sslkey" {
  sensitive = true
  type      = string
}
variable "allow_insecure_connections" { type = bool }
variable "read_only" { type = bool }
