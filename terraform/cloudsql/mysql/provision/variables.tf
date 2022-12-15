variable "authorized_network" { type = string }
variable "authorized_network_id" { type = string }
variable "instance_name" { type = string }
variable "db_name" { type = string }

variable "region" { type = string }
variable "labels" { type = map(any) }
variable "storage_gb" { type = number }
variable "mysql_version" { type = string }

variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }
variable "tier" { type = string }
variable "disk_autoresize" { type = bool }
variable "disk_autoresize_limit" { type = number }
variable "deletion_protection" { type = bool }
variable "backups_start_time" { type = string }
variable "backups_location" {
  type    = string
  default = null
}
variable "backups_retain_number" { type = number }
variable "backups_transaction_log_retention_days" { type = number }