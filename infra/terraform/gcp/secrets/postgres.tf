resource "google_secret_manager_secret" "postgres_creds" {
  secret_id = "postgres-creds"   # ← FIXED NAME

  labels = merge(local.common_labels, {
    service = "postgres"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "postgres_creds" {
  secret      = google_secret_manager_secret.postgres_creds.id
  secret_data = var.postgres_app_password
}
