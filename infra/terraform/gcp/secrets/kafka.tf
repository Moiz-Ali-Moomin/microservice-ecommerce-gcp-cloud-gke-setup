resource "google_secret_manager_secret" "kafka_password" {
  secret_id = "kafka-password"
  labels = merge(local.common_labels, {
    service = "kafka"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "kafka_password" {
  secret      = google_secret_manager_secret.kafka_password.id
  secret_data = var.kafka_password
}
