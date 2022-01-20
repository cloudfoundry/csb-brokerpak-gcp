variable "worker_machine_type" { type = string }
variable "master_machine_type" { type = string }
variable "worker_count" { type = number }
variable "master_count" { type = number }
variable "preemptible_count" { type = number }

variable "name" { type = string }
variable "region" { type = string }
variable "labels" { type = map(any) }

variable "credentials" { type = string }
variable "project" { type = string }
