output "id" { value = google_storage_bucket.bucket.id }
output "bucket_name" { value = var.name }
output "status" { value = format("created bucket %s - URL: https://console.cloud.google.com/storage/browser/%s;tab=objects?project=%s",
  google_storage_bucket.bucket.name,
  google_storage_bucket.bucket.name, var.project
) }
