
resource "google_service_account" "account" {
  account_id   = var.service_account_name
  display_name = format("%s with role roles/%s", var.service_account_display_name, var.role)
}

resource "google_service_account_key" "key" {
  service_account_id = google_service_account.account.name
}

resource "google_pubsub_topic_iam_member" "member" {
  project = var.project
  topic   = var.topic_name
  role    = format("roles/%s", var.role)
  member  = format("serviceAccount:%s", google_service_account.account.email)
}

resource "google_pubsub_subscription_iam_member" "member" {
  count        = length(var.subscription_name) > 0 && var.role != "pubsub.publisher" ? 1 : 0
  subscription = var.subscription_name
  role         = format("roles/%s", var.role)
  member       = format("serviceAccount:%s", google_service_account.account.email)
}