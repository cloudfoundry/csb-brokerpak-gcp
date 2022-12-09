variable "authorized_network" { type = string }
variable "authorized_network_id" { type = string }
variable "instance_name" { type = string }
variable "db_name" { type = string }

variable "region" { type = string }
variable "labels" { type = map(any) }
variable "storage_gb" { type = number }
variable "database_version" { type = string }

variable "credentials" { type = string }
variable "project" { type = string }
variable "tier" { type = string }
variable "disk_autoresize" { type = bool }
variable "disk_autoresize_limit" { type = number }
