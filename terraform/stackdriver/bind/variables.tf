variable "name" { type = string }
variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }
variable "role" { type = string }
