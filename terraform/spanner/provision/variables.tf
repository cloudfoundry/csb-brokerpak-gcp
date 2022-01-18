variable "credentials" { type = string }
variable "project" { type = string }
variable "labels" { type = map(any) }
variable "ddl" { type = list(any) }
variable "num_nodes" { type = string }
variable "instance_name" { type = string }
variable "config" { type = string }
