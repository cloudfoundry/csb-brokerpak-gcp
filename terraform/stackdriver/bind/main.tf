resource "google_service_account" "account" {
  account_id   = substr(var.name, 0, 30)
  display_name = format("%s with role %s", var.name, var.role)
}

resource "google_service_account_key" "key" {
  service_account_id = google_service_account.account.name
}

resource "google_project_iam_member" "member" {
  project = var.project
  role    = format("roles/%s", var.role)
  member  = format("serviceAccount:%s", google_service_account.account.email)
}
