resource "google_secret_manager_secret" "api_gateway_postgres_dsn" {
  secret_id = "api-gateway-postgres-dsn"
  labels = merge(local.common_labels, {
    service = "api-gateway"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "api_gateway_postgres_dsn" {
  secret      = google_secret_manager_secret.api_gateway_postgres_dsn.id
  secret_data = var.api_gateway_postgres_dsn
}
