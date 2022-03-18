output "name" { value = google_sql_database.database.name }
output "hostname" { value = google_sql_database_instance.instance.first_ip_address }

output "port" { value = var.db_port }
output "username" { value = google_sql_user.admin_user.name }
output "password" {
  sensitive = true
  value     = google_sql_user.admin_user.password
}
output "use_tls" { value = false }
