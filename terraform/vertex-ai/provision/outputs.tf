output "project_id" {
  value = local.resolved_project
}

output "instance_name" {
  value = var.instance_name
}

output "cf_provenance_json" {
  value = jsonencode(local.cf_provenance)
}

output "region" {
  value = var.region
}

output "resource_labels_json" {
  value = jsonencode(local.common_labels)
}

output "service_account_email" {
  value = google_service_account.vertex_ai.email
}

output "service_account_id" {
  value = google_service_account.vertex_ai.id
}

output "bucket_name" {
  value = google_storage_bucket.ml_artifacts.name
}

output "api_endpoint" {
  value = "${var.region}-aiplatform.googleapis.com"
}

output "available_models" {
  value = var.models
}

output "budget_amount" {
  value = var.budget_amount
}

output "budget_enforcement_mode" {
  value = local.budget_enforcement_mode
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}
