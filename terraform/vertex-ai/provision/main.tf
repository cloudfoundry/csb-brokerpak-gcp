# Enable Vertex AI API
resource "google_project_service" "vertex_ai" {
  project            = local.resolved_project
  service            = "aiplatform.googleapis.com"
  disable_on_destroy = false
}

# Enable Cloud Storage API (needed for the artifact bucket)
resource "google_project_service" "storage" {
  project            = local.resolved_project
  service            = "storage.googleapis.com"
  disable_on_destroy = false
}

# Dedicated service account for Vertex AI workloads
resource "google_service_account" "vertex_ai" {
  account_id   = trimsuffix(substr("csb-vai-${var.instance_name}", 0, 30), "-")
  display_name = "CSB Vertex AI — ${var.instance_name}"
  project      = local.resolved_project
  description  = "Sandbox service account. Expires ${local.ttl_expires_at}."
}

# Grant Vertex AI User role on the project
resource "google_project_iam_member" "vertex_user" {
  project = local.resolved_project
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${google_service_account.vertex_ai.email}"
}

# Grant Storage Object Admin on the sandbox artifact bucket only
resource "google_storage_bucket_iam_member" "sa_bucket_admin" {
  bucket = google_storage_bucket.ml_artifacts.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.vertex_ai.email}"
}

# GCS bucket for ML artifacts — auto-deletes objects after 10 days
resource "google_storage_bucket" "ml_artifacts" {
  name                        = "${local.resolved_project}-csb-vai-${substr(var.instance_name, 0, 20)}"
  location                    = var.region
  storage_class               = "STANDARD"
  project                     = local.resolved_project
  uniform_bucket_level_access = true
  labels                      = local.common_labels

  lifecycle_rule {
    action { type = "Delete" }
    condition { age = 10 }
  }

  depends_on = [google_project_service.storage]
}
