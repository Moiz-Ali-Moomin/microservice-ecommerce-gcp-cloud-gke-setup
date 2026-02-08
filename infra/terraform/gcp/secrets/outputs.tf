output "postgres_secret_id" {
  description = "Secret ID for Postgres App Password"
  value       = google_secret_manager_secret.postgres_app_password.secret_id
}

output "redis_secret_id" {
  description = "Secret ID for Redis Password"
  value       = google_secret_manager_secret.redis_password.secret_id
}

output "kafka_secret_id" {
  description = "Secret ID for Kafka Password"
  value       = google_secret_manager_secret.kafka_password.secret_id
}

output "airflow_secret_id" {
  description = "Secret ID for Airflow DB Password"
  value       = google_secret_manager_secret.airflow_db_password.secret_id
}

output "metabase_secret_id" {
  description = "Secret ID for Metabase DB Password"
  value       = google_secret_manager_secret.metabase_db_password.secret_id
}
