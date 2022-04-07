output "name" { value = google_sql_database.database.name }
output "hostname" { value = google_sql_database_instance.instance.first_ip_address }

output "username" { value = postgresql_role.createrole_user.name }
output "password" {
  sensitive = true
  value     = postgresql_role.createrole_user.password
}
output "use_tls" { value = false }
