resource "google_secret_manager_secret" "redis_secret" {
  secret_id = "redis-secret"   # ← FIXED NAME

  labels = merge(local.common_labels, {
    service = "redis"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "redis_secret" {
  secret      = google_secret_manager_secret.redis_secret.id
  secret_data = var.redis_password
}
