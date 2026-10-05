locals {
  # project is always provided via computed_input from gcp.project config.
  resolved_project        = var.project
  ttl_expires_at          = timeadd(timestamp(), "${var.ttl_hours}h")
  budget_enforcement_mode = trimspace(var.budget_alert_email) == "" ? "alert-email-optional-unset" : "alert-email-configured"
  cf_context               = try(jsondecode(var.cf_context_json), {})
  cf_originating_identity = try(jsondecode(var.cf_originating_identity_json), {})
  cf_user_id              = try(local.cf_originating_identity.user_id, local.cf_originating_identity.value.user_id, "")

  cf_provenance = {
    cf_organization_guid = try(local.cf_context.organization_guid, "")
    cf_organization_name = try(local.cf_context.organization_name, "")
    cf_space_guid        = try(local.cf_context.space_guid, "")
    cf_space_name        = try(local.cf_context.space_name, "")
    cf_user_id           = local.cf_user_id
  }

  cf_provenance_labels = {
    for key, value in local.cf_provenance : key => substr(lower(value), 0, 63)
    if trimspace(value) != ""
  }

  common_labels = merge(var.labels, local.cf_provenance_labels, {
    ttlexpiry   = lower(replace(local.ttl_expires_at, ":", "-"))
    managed-by  = "cloud-service-broker"
    environment = "sandbox"
  })
}
