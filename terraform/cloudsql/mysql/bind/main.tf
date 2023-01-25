resource "csbmysql_binding_user" "new_user" {
  username = random_string.username.result
  password = random_password.password.result
}