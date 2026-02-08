resource "google_artifact_registry_repository" "repo" {
  location      = var.region
  repository_id = var.registry_id
  format        = "DOCKER"

  labels = {
    env  = "prod"
    team = "platform"
  }
  lifecycle {
    prevent_destroy = true
  }
}

