resource "random_string" "username" {
  length  = 16
  special = false
  numeric = false
}

resource "random_password" "password" {
  length           = 64
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "csbmysql_binding_user" "new_user" {
  username                   = random_string.username.result
  password                   = random_password.password.result
  allow_insecure_connections = var.allow_insecure_connections
  read_only                  = var.read_only
}