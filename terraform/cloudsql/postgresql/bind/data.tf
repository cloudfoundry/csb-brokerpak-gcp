locals {
  roles_map    = jsondecode(var.custom_roles)
  custom_roles = try(local.roles_map[var.app_guid], local.roles_map["*"], "binding_user_group")
  actual_roles = var.enable_custom_roles ? local.custom_roles : "binding_user_group"
}
