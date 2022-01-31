output "policy_etag" { value = join(",", google_spanner_database_iam_policy.spanner_database_iam_policy.*.etag) }
output "binding_etag" { value = join(",", google_spanner_database_iam_binding.spanner_database_iam_binding.*.etag) }
output "member_etag" { value = join(",", google_spanner_database_iam_member.spanner_database_iam_member.*.etag) }
output "Name" { value = google_service_account.account.name }
output "Email" { value = google_service_account.account.email }
output "UniqueId" { value = google_service_account.account.unique_id }
output "PrivateKeyData" {
  sensitive = true
  value     = google_service_account_key.key.private_key
}
output "ProjectId" { value = google_service_account.account.project }
output "instance" { value = var.instance }
output "db_name" { value = var.db_name }
output "Credentials" {
  sensitive = true
  value     = base64decode(google_service_account_key.key.private_key)
}
