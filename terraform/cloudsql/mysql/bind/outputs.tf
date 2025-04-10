output "username" { value = csbmysql_binding_user.new_user.username }
output "password" {
  value     = csbmysql_binding_user.new_user.password
  sensitive = true
}
output "uri" {
  sensitive = true
  value = format("mysql://%s:%s@%s:%d/%s",
    random_string.username.result,
    random_password.password.result,
    var.mysql_hostname,
    local.port,
  var.mysql_db_name)
}
output "port" { value = local.port }
output "jdbcUrl" {
  sensitive = true
  value = format("jdbc:mysql://%s:%d/%s?user=%s\u0026password=%s\u0026useSsl=%s%s",
    var.mysql_hostname,
    local.port,
    var.mysql_db_name,
    csbmysql_binding_user.new_user.username,
    csbmysql_binding_user.new_user.password,
    local.useSSL,
    local.jdbcUrlSuffix
  )
}
output "sslrootcert" { value = var.sslrootcert }
output "sslcert" { value = var.sslcert }
output "sslkey" {
  sensitive = true
  value     = var.sslkey
}
output "read_only" { value = var.read_only }