output "credentials_json" {
  value     = base64decode(google_service_account_key.binding_key.private_key)
  sensitive = true
}

output "instance_name" {
  value = var.instance_name
}

output "resource_labels_json" {
  value = var.resource_labels_json
}

output "binding_provenance_json" {
  value = jsonencode({
    cf_organization_guid = try(jsondecode(var.cf_context_json).organization_guid, "")
    cf_organization_name = try(jsondecode(var.cf_context_json).organization_name, "")
    cf_space_guid        = try(jsondecode(var.cf_context_json).space_guid, "")
    cf_space_name        = try(jsondecode(var.cf_context_json).space_name, "")
    cf_user_id           = try(jsondecode(var.cf_originating_identity_json).user_id, jsondecode(var.cf_originating_identity_json).value.user_id, "")
    cf_app_guid          = var.cf_app_guid
  })
}

output "project_id" {
  value = var.project_id
}

output "region" {
  value = var.region
}

output "api_endpoint" {
  value = var.api_endpoint
}

output "models" {
  value = var.available_models
}

output "bucket_name" {
  value = var.bucket_name
}

output "budget_enforcement_mode" {
  value = var.budget_enforcement_mode
}

output "ttl_expires_at" {
  value = var.ttl_expires_at
}

output "normalized_binding_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "gcp"
    provisioner_family = "google_vertex_identity"
    connection_type    = "runtime"
    endpoint = {
      base_url    = var.api_endpoint
      region      = var.region
      api_version = null
    }
    access = {
      mode       = "gcp_access_token"
      expires_at = null
    }
    grant = {
      kind                 = "service_account"
      least_privilege_unit = "project"
      allowed_models       = try(jsondecode(var.available_models), [])
    }
    credential = {
      format = "service_account_json"
      inline = {
        credentials_json = base64decode(google_service_account_key.binding_key.private_key)
      }
      secret_ref = null
    }
  })
  sensitive = true
}
