resource "google_secret_manager_secret" "postgres_app_password" {
  secret_id = "postgres-app-password"
  labels = merge(local.common_labels, {
    service = "postgres"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "postgres_app_password" {
  secret      = google_secret_manager_secret.postgres_app_password.id
  secret_data = var.postgres_app_password
}
