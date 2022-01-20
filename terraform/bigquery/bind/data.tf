locals {
  members = format("serviceAccount:%s", google_service_account.account.email)
}

data "google_iam_policy" "database_iam_policy" {
  binding {
    role    = var.role
    members = [local.members]
  }
}
