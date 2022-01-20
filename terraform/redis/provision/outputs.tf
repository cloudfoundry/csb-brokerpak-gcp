output "memory_size_gb" { value = google_redis_instance.instance.memory_size_gb }
output "service_tier" { value = google_redis_instance.instance.tier }
output "redis_version" { value = google_redis_instance.instance.redis_version }
output "host" { value = google_redis_instance.instance.host }
output "port" { value = google_redis_instance.instance.port }
