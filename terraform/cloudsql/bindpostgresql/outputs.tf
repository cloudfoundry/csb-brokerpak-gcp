output "username" { value = random_string.username.result }
output "password" { value = random_password.password.result }
output "uri" {
  value = format("postgresql://%s:%s@%s:%d/%s",
    random_string.username.result,
    random_password.password.result,
    var.hostname,
    var.port,
  var.db_name)
}
/* output jdbcUrl {
  value = format("jdbc:postgresql://%s:%d/%s?user=%s\u0026password=%s\u0026useSSL=false",
                  var.hostname,
                  var.port,
                  var.db_name,
                  random_string.username.result,
                  random_password.password.result)
} */

output "jdbcUrl" {
  value = format("jdbc:%s://%s:%s/%s?user=%s\u0026password=%s\u0026verifyServerCertificate=true\u0026useSSL=%v\u0026requireSSL=false",
    "postgresql",
    var.hostname,
    var.port,
    var.db_name,
    random_string.username.result,
    random_password.password.result,
  var.use_tls)
}
