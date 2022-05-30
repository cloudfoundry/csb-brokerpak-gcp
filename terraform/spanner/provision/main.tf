#---------------------------------------------------
# Create google spanner instance
#---------------------------------------------------
resource "google_spanner_instance" "spanner_instance" {

  config       = var.num_nodes > 2 ? var.config : "regional-${var.config}"
  display_name = local.display_name
  name         = lower(var.instance_name)
  num_nodes    = var.num_nodes
  project      = var.project
  labels       = var.labels

  lifecycle {
    ignore_changes        = []
    create_before_destroy = true
    prevent_destroy       = true
  }
}

#---------------------------------------------------
# Create spanner database
#---------------------------------------------------
resource "google_spanner_database" "spanner_database" {


  instance = google_spanner_instance.spanner_instance.name
  name     = local.db_name
  project  = var.project

  ddl = var.ddl

  lifecycle {
    ignore_changes        = []
    create_before_destroy = true
  }

  deletion_protection = false
}
