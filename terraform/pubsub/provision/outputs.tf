output "topic_name" { value = google_pubsub_topic.topic.name }
output "subscription_name" { value = var.subscription_name }
output "status" { value = format("topic %s is created", google_pubsub_topic.topic.name) }
