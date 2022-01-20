locals {
  display_name = substr(var.instance_name, 4, 29)
  db_name      = replace(local.display_name, "-", "_")
}
