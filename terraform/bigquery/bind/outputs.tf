output "Name" { value = google_service_account.account.name }
output "Email" { value = google_service_account.account.email }
output "UniqueId" { value = google_service_account.account.unique_id }
output "PrivateKeyData" {
  sensitive = true
  value     = google_service_account_key.key.private_key
}
output "ProjectId" { value = google_service_account.account.project }
output "dataset_id" { value = var.dataset_id }
output "Credentials" {
  sensitive = true
  value     = base64decode(google_service_account_key.key.private_key)
}
