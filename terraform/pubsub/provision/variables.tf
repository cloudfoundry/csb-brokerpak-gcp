variable "topic_name" { type = string }
variable "subscription_name" { type = string }
variable "ack_deadline" { type = number }
variable "labels" { type = map(any) }
variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }