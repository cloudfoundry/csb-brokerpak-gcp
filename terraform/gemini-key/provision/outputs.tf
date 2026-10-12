output "project_id" {
  value = local.resolved_project
}

output "instance_name" {
  value = var.instance_name
}

output "cf_provenance_json" {
  value = jsonencode(local.cf_provenance)
}

output "resource_labels_json" {
  value = jsonencode(local.common_labels)
}

output "api_endpoint" {
  value = "https://generativelanguage.googleapis.com"
}

output "api_key" {
  value     = google_apikeys_key.gemini.key_string
  sensitive = true
}

output "available_models" {
  value = var.models
}

output "api_key_resource_name" {
  value = google_apikeys_key.gemini.id
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