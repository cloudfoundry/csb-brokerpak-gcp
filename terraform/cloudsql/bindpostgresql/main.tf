resource "random_string" "username" {
  length  = 16
  special = false
  number  = false
}

resource "random_password" "password" {
  length           = 64
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "postgresql_role" "new_user" {
  name                = random_string.username.result
  login               = true
  password            = random_password.password.result
  skip_reassign_owned = true
  skip_drop_role      = true
}

resource "postgresql_grant" "db_access" {
  depends_on  = [postgresql_role.new_user]
  database    = var.db_name
  role        = postgresql_role.new_user.name
  object_type = "database"
  privileges  = ["ALL"]
}

resource "postgresql_grant" "table_access" {
  depends_on  = [postgresql_role.new_user]
  database    = var.db_name
  role        = postgresql_role.new_user.name
  schema      = "public"
  object_type = "table"
  privileges  = ["ALL"]
}
