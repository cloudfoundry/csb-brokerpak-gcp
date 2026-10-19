resource "google_project_service" "generative_language" {
  project            = local.resolved_project
  service            = "generativelanguage.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "apikeys" {
  project            = local.resolved_project
  service            = "apikeys.googleapis.com"
  disable_on_destroy = false
}

resource "google_apikeys_key" "gemini" {
  project      = local.resolved_project
  name         = substr("csb-gemini-${var.instance_name}", 0, 63)
  display_name = "CSB Gemini API — ${var.instance_name}"

  restrictions {
    api_targets {
      service = "generativelanguage.googleapis.com"
    }
  }

  depends_on = [
    google_project_service.generative_language,
    google_project_service.apikeys,
  ]
}