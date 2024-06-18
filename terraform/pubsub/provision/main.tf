resource "google_pubsub_topic" "topic" {
  name                       = var.topic_name
  message_retention_duration = var.topic_message_retention_duration
  kms_key_name               = var.topic_kms_key_name
  labels                     = var.labels
}

resource "google_pubsub_subscription" "subscription" {
  count = length(var.subscription_name) > 0 ? 1 : 0

  name                         = var.subscription_name
  topic                        = google_pubsub_topic.topic.name
  ack_deadline_seconds         = var.ack_deadline
  retain_acked_messages        = var.subscription_retain_acked_messages
  message_retention_duration   = var.subscription_message_retention_duration
  enable_message_ordering      = var.subscription_enable_message_ordering
  enable_exactly_once_delivery = var.subscription_enable_exactly_once_delivery

  dynamic "push_config" {
    for_each = length(var.push_endpoint) > 0 ? [1] : []
    content {
      push_endpoint = var.push_endpoint
    }
  }

  dynamic "expiration_policy" {
    for_each = length(var.subscription_expiration_policy) > 0 ? [1] : []
    content {
      ttl = var.subscription_expiration_policy
    }
  }

  dynamic "retry_policy" {
    for_each = length(var.subscription_retry_policy_minimum_backoff) > 0 && length(var.subscription_retry_policy_maximum_backoff) > 0 ? [1] : []
    content {
      minimum_backoff = var.subscription_retry_policy_minimum_backoff
      maximum_backoff = var.subscription_retry_policy_maximum_backoff
    }
  }

  labels = var.labels
}

