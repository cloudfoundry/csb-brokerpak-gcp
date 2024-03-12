variable "service_tier" { type = string }
variable "authorized_network" { type = string }
variable "authorized_network_id" { type = string }
variable "display_name" { type = string }
variable "instance_id" { type = string }
variable "region" { type = string }
variable "memory_size_gb" { type = number }
variable "labels" { type = map(any) }
variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }
variable "reserved_ip_range" { type = string }
