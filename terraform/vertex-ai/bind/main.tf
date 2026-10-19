# Per-binding service account key for the provisioned Vertex AI service account.
resource "google_service_account_key" "binding_key" {
  service_account_id = var.service_account_id
  key_algorithm      = "KEY_ALG_RSA_2048"
}
