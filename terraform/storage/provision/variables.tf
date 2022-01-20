variable "name" { type = string }
variable "region" { type = string }
variable "storage_class" { type = string }
variable "labels" { type = map(any) }
variable "acl" { type = string }
variable "credentials" { type = string }
variable "project" { type = string }
