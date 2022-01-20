resource "google_service_account" "account" {
  account_id   = var.service_account_name
  display_name = var.service_account_display_name
}
resource "google_service_account_key" "key" {
  service_account_id = google_service_account.account.name
}

resource "google_bigquery_dataset_access" "access" {
  project       = var.project
  dataset_id    = var.dataset_id
  role          = "roles/${var.role}"
  user_by_email = google_service_account.account.email
}
