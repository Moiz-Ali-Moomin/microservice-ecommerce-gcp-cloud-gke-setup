resource "google_artifact_registry_repository" "repo" {
  location      = var.region
  repository_id = "ecommerce-repo"
  format        = "DOCKER"

  labels = {
    env  = "prod"
    team = "platform"
  }
}
