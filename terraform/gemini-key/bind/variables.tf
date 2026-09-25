variable "instance_name" { type = string }
variable "resource_labels_json" { type = string }
variable "project_id" { type = string }
variable "api_endpoint" { type = string }
variable "api_key" {
  type      = string
  sensitive = true
}
variable "available_models" { type = string }
variable "budget_enforcement_mode" { type = string }
variable "ttl_expires_at" { type = string }
variable "cf_context_json" { type = string }
variable "cf_originating_identity_json" { type = string }
variable "cf_app_guid" {
  type    = string
  default = ""
}