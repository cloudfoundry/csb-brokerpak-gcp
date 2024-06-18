variable "topic_name" { type = string }
variable "subscription_name" { type = string }
variable "ack_deadline" { type = number }
variable "push_endpoint" { type = string }
variable "topic_message_retention_duration" { type = string }
variable "topic_kms_key_name" { type = string }
variable "subscription_message_retention_duration" { type = string }
variable "subscription_retain_acked_messages" { type = bool }
variable "subscription_expiration_policy" { type = string }
variable "subscription_retry_policy_minimum_backoff" { type = string }
variable "subscription_retry_policy_maximum_backoff" { type = string }
variable "subscription_enable_message_ordering" { type = bool }
variable "subscription_enable_exactly_once_delivery" { type = bool }
variable "labels" { type = map(any) }
variable "credentials" {
  type      = string
  sensitive = true
}
variable "project" { type = string }
