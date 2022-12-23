variable "tier" { type = string }
variable "authorized_network" { type = string }
variable "authorized_network_id" { type = string }
variable "authorized_networks_cidrs" { type = list(string) }
variable "public_ip" { type = bool }
variable "instance_name" { type = string }
variable "db_name" { type = string }

variable "region" { type = string }
variable "labels" { type = map(any) }
variable "storage_gb" { type = number }
variable "database_version" { type = string }
variable "backups_retain_number" { type = number }
variable "backups_location" { type = string }
variable "backups_start_time" { type = string }
variable "backups_point_in_time_log_retain_days" { type = number }

variable "credentials" { type = string }
variable "project" { type = string }
variable "require_ssl" { type = bool }
