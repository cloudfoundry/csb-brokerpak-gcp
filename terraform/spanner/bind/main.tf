resource "google_service_account" "account" {
  account_id   = var.service_account_name
  display_name = var.service_account_display_name
}
resource "google_service_account_key" "key" {
  service_account_id = google_service_account.account.name
}

resource "google_spanner_database_iam_policy" "spanner_database_iam_policy" {

  instance    = var.instance
  database    = var.db_name
  policy_data = data.google_iam_policy.database_iam_policy.policy_data

  depends_on = [data.google_iam_policy.database_iam_policy]

  lifecycle {
    ignore_changes        = []
    create_before_destroy = true
  }
}

resource "google_spanner_database_iam_binding" "spanner_database_iam_binding" {
  instance = var.instance
  database = var.db_name
  role     = var.role

  members = [local.members]

  lifecycle {
    ignore_changes        = []
    create_before_destroy = true
  }
}

resource "google_spanner_database_iam_member" "spanner_database_iam_member" {
  instance = var.instance
  database = var.db_name
  role     = var.role
  member   = local.members

  lifecycle {
    ignore_changes        = []
    create_before_destroy = true
  }
}
