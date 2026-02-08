resource "google_secret_manager_secret" "metabase_db_password" {
  secret_id = "metabase-db-password"
  labels = merge(local.common_labels, {
    service = "metabase"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "metabase_db_password" {
  secret      = google_secret_manager_secret.metabase_db_password.id
  secret_data = var.metabase_db_password
}
