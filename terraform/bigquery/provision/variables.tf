variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }
variable "labels" { type = map(any) }
variable "region" { type = string }
variable "instance_name" { type = string }
