output "name" { value = google_sql_database.database.name }
output "hostname" { value = google_sql_database_instance.instance.first_ip_address }

output "username" { value = google_sql_user.admin_user.name }
output "password" {
  sensitive = true
  value     = google_sql_user.admin_user.password
}
output "require_ssl" { value = var.require_ssl }

output "sslcert" { value = google_sql_ssl_cert.client_cert.cert }
output "sslkey" {
  value     = google_sql_ssl_cert.client_cert.private_key
  sensitive = true
}
output "sslrootcert" { value = google_sql_database_instance.instance.server_ca_cert.0.cert }