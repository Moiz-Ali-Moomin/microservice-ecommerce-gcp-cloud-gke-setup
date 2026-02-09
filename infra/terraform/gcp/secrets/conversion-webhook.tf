resource "google_secret_manager_secret" "clickbank_secret_key" {
  secret_id = "clickbank-secret-key"
  labels = merge(local.common_labels, {
    service = "conversion-webhook"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "clickbank_secret_key" {
  secret      = google_secret_manager_secret.clickbank_secret_key.id
  secret_data = var.clickbank_secret_key
}
