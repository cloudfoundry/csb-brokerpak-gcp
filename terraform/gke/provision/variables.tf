variable "cluster_name" {
  type = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{0,38}[a-z0-9]$", var.cluster_name))
    error_message = "cluster_name must be a valid GKE cluster name with at most 40 characters."
  }
}

variable "zone" {
  type = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]+-[a-z]$", var.zone))
    error_message = "zone must be a valid GCP zone name."
  }
}

variable "network" {
  type = string

  validation {
    condition     = trimspace(var.network) == "" || can(regex("^[a-z][a-z0-9-]{0,61}[a-z0-9]$", var.network))
    error_message = "network must be an existing VPC network name in the configured project."
  }
}

variable "subnetwork" {
  type = string

  validation {
    condition     = trimspace(var.subnetwork) == "" || can(regex("^[a-z][a-z0-9-]{0,61}[a-z0-9]$", var.subnetwork))
    error_message = "subnetwork must be empty or an existing subnetwork name in the configured project."
  }
}

variable "project" { type = string }
variable "ttl_hours" { type = number }
variable "labels" { type = map(any) }
variable "credentials" {
  type      = string
  sensitive = true
}
