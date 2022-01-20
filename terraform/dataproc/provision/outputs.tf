output "bucket_name" { value = google_dataproc_cluster.cluster.cluster_config.0.bucket }
output "cluster_name" { value = google_dataproc_cluster.cluster.name }
output "region" { value = google_dataproc_cluster.cluster.region }
