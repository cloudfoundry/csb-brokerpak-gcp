variable "name" { type = string }
variable "region" { type = string }
variable "storage_class" { type = string }
variable "labels" { type = map(any) }
variable "credentials" { type = string }
variable "project" { type = string }
variable "placement_dual_region_data_locations" { type = list(string) }
