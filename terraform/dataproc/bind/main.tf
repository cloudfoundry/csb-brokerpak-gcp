resource "google_service_account" "account" {
  account_id   = var.service_account_name
  display_name = var.service_account_name
}

resource "google_service_account_key" "key" {
  service_account_id = google_service_account.account.name
}

resource "google_storage_bucket_iam_member" "member" {
  bucket = var.bucket
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.account.email}"
}

resource "google_project_iam_member" "member" {
  project = var.project
  role    = "roles/dataproc.editor"
  member  = "serviceAccount:${google_service_account.account.email}"
}
