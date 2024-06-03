output "name" { value = google_service_account.account.name }
output "email" { value = google_service_account.account.email }
output "UniqueId" { value = google_service_account.account.unique_id }
output "PrivateKeyData" {
  sensitive = true
  value     = google_service_account_key.key.private_key
}
output "ProjectId" { value = google_service_account.account.project }
output "credentials" {
  sensitive = true
  value     = base64decode(google_service_account_key.key.private_key)
}

// Legacy outputs
//      Name:           newSA.DisplayName,
//		Email:          newSA.Email,
//		UniqueId:       newSA.UniqueId,
//		PrivateKeyData: newSAKey.PrivateKeyData,
//		ProjectId:      sam.ProjectId,