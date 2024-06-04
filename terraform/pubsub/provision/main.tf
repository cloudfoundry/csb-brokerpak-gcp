resource "google_pubsub_topic" "topic" {
  name   = var.topic_name
  labels = var.labels
}

resource "google_pubsub_subscription" "subscription" {
  count = length(var.subscription_name) > 0 ? 1 : 0

  name                 = var.subscription_name
  topic                = google_pubsub_topic.topic.name
  ack_deadline_seconds = var.ack_deadline
  dynamic "push_config" {
    for_each = length(var.push_endpoint) > 0 ? [1] : []
    content {
      push_endpoint = var.push_endpoint
    }
  }

  labels = var.labels
}

