variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }
variable "role" { type = string }
variable "service_account_name" { type = string }
variable "service_account_display_name" { type = string }
variable "bucket" { type = string }
