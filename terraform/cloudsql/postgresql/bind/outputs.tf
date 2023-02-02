output "username" { value = random_string.username.result }
output "password" {
  sensitive = true
  value     = random_password.password.result
}
output "uri" {
  sensitive = true
  value = format("postgresql://%s:%s@%s:%d/%s",
    random_string.username.result,
    random_password.password.result,
    var.hostname,
    local.port,
    var.db_name,
  )
}
output "port" { value = local.port }
output "jdbcUrl" {
  sensitive = true
  value = format("jdbc:postgresql://%s:%s/%s?user=%s\u0026password=%s\u0026ssl=%v",
    var.hostname,
    local.port,
    var.db_name,
    random_string.username.result,
    random_password.password.result,
    var.require_ssl,
  )
}
output "name" { value = var.db_name }
output "hostname" { value = var.hostname }
output "require_ssl" { value = var.require_ssl }
output "sslcert" { value = var.sslcert }
output "sslkey" { value = var.sslkey }
output "sslrootcert" { value = var.sslrootcert }
output "private_ip" { value = var.private_ip }

