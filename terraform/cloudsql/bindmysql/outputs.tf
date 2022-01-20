output "username" { value = mysql_user.newuser.user }
output "password" { value = random_password.password.result }
output "uri" {
  value = format("mysql://%s:%s@%s:%d/%s",
    random_string.username.result,
    random_password.password.result,
    var.mysql_hostname,
    var.mysql_port,
  var.mysql_db_name)
}
output "jdbcUrl" {
  value = format("jdbc:mysql://%s:%d/%s?user=%s\u0026password=%s\u0026useSSL=%v",
    var.mysql_hostname,
    var.mysql_port,
    var.mysql_db_name,
    mysql_user.newuser.user,
    random_password.password.result,
  var.use_tls)
}
