resource "google_secret_manager_secret" "airflow_db_password" {
  secret_id = "airflow-db-password"
  labels = merge(local.common_labels, {
    service = "airflow"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "airflow_db_password" {
  secret      = google_secret_manager_secret.airflow_db_password.id
  secret_data = var.airflow_db_password
}
