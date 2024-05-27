output "name" { value = google_service_account.account.name }
output "email" { value = google_service_account.account.email }
output "unique_id" { value = google_service_account.account.unique_id }
output "PrivateKeyData" {
  sensitive = true
  value     = google_service_account_key.key.private_key
}
output "ProjectId" { value = google_service_account.account.project }
output "private_key_data" {
  sensitive   = true
  value       = google_service_account_key.key.private_key
  description = "Deprecated - The private key data of the service account"
}
output "project_id" {
  value       = google_service_account.account.project
  description = "Deprecated - The project ID of the service account"
}
output "credentials" {
  sensitive = true
  value     = base64decode(google_service_account_key.key.private_key)
}
