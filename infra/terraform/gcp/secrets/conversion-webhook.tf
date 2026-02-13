resource "google_secret_manager_secret" "conversion_webhook" {
  secret_id = "conversion-webhook"   # ← FIXED NAME

  labels = merge(local.common_labels, {
    service = "conversion-webhook"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "conversion_webhook" {
  secret      = google_secret_manager_secret.conversion_webhook.id
  secret_data = var.clickbank_secret_key
}
