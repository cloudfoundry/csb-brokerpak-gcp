variable "name" { type = string }
variable "region" { type = string }
variable "storage_class" { type = string }
variable "labels" { type = map(any) }
variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }
variable "placement_dual_region_data_locations" { type = list(string) }
variable "versioning" { type = bool }
variable "public_access_prevention" { type = string }
variable "uniform_bucket_level_access" { type = bool }
variable "default_kms_key_name" { type = string }
variable "autoclass" { type = bool }
variable "retention_policy_is_locked" { type = bool }
variable "retention_policy_retention_period" { type = number }
variable "predefined_acl" { type = string }