output "email" { value = google_service_account.account.email }
output "private_key" {
  sensitive = true
  value     = google_service_account_key.key.private_key
}
output "project_id" { value = google_service_account.account.project }
output "name" { value = google_service_account.account.account_id }
