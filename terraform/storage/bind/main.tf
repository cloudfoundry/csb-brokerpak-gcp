
resource "google_service_account" "account" {
  account_id   = var.service_account_name
  display_name = var.service_account_display_name
}
resource "google_service_account_key" "key" {
  service_account_id = google_service_account.account.name
}
resource "google_storage_bucket_iam_member" "member" {
  bucket = var.bucket
  role   = format("roles/%s", var.role)
  member = format("serviceAccount:%s", google_service_account.account.email)
}
