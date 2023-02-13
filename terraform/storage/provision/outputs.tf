output "id" { value = google_storage_bucket.bucket.id }
output "bucket_name" { value = var.name }
output "status" { value = format("service %s created", google_storage_bucket.bucket.name) }
