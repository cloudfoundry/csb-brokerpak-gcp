output "instance" { value = google_spanner_instance.spanner_instance.name }
output "db_name" { value = join(",", google_spanner_database.spanner_database.*.name) }
