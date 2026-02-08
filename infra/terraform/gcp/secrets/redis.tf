resource "google_secret_manager_secret" "redis_password" {
  secret_id = "redis-password"
  labels = merge(local.common_labels, {
    service = "redis"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "redis_password" {
  secret      = google_secret_manager_secret.redis_password.id
  secret_data = var.redis_password
}
