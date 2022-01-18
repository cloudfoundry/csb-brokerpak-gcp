data "google_iam_policy" "database_iam_policy" {
  binding {
    role    = var.role
    members = [local.members]
  }
}
