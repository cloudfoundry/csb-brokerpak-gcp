

resource "google_bigquery_dataset" "csb_dataset" {
  dataset_id    = replace(var.instance_name, "-", "")
  friendly_name = var.instance_name
  location      = var.region
  access {
    role          = "OWNER"
    special_group = "projectOwners"
  }
  access {
    role          = "WRITER"
    special_group = "projectWriters"
  }
  access {
    role          = "READER"
    special_group = "allAuthenticatedUsers"
  }

  lifecycle {
    prevent_destroy = true
  }
}
