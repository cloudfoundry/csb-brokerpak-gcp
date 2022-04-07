output "name" { value = google_sql_database.database.name }
output "hostname" { value = google_sql_database_instance.instance.first_ip_address }

output "username" { value = postgresql_role.createrole_user.name }
output "password" {
  sensitive = true
  value     = postgresql_role.createrole_user.password
}
output "use_tls" { value = var.use_tls }

output "sslcert" { value = google_sql_ssl_cert.client_cert.cert }
output "sslkey" {
    value = google_sql_ssl_cert.client_cert.private_key
    sensitive   = true
}
output "sslrootcert" { value = google_sql_database_instance.instance.server_ca_cert.0.cert }
