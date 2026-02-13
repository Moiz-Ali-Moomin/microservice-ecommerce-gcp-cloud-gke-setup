output "postgres_secret_id" {
  description = "Secret ID for Postgres credentials"
  value       = google_secret_manager_secret.postgres_creds.secret_id
}

output "redis_secret_id" {
  description = "Secret ID for Redis credentials"
  value       = google_secret_manager_secret.redis_secret.secret_id
}

output "airflow_secret_id" {
  description = "Secret ID for Airflow DB password"
  value       = google_secret_manager_secret.airflow_db_password.secret_id
}

output "metabase_secret_id" {
  description = "Secret ID for Metabase DB password"
  value       = google_secret_manager_secret.metabase_db_password.secret_id
}

output "kafka_secret_id" {
  description = "Secret ID for Kafka password"
  value       = google_secret_manager_secret.kafka_password.secret_id
}

output "auth_jwt_secret_id" {
  description = "Secret ID for Auth JWT"
  value       = google_secret_manager_secret.auth_service_jwt.secret_id
}

output "auth_db_secret_id" {
  description = "Secret ID for Auth DB"
  value       = google_secret_manager_secret.auth_service_db.secret_id
}

output "auth_redis_secret_id" {
  description = "Secret ID for Auth Redis"
  value       = google_secret_manager_secret.auth_service_redis.secret_id
}

output "api_gateway_secret_id" {
  description = "Secret ID for API Gateway DB"
  value       = google_secret_manager_secret.api_gateway_postgres_dsn.secret_id
}

output "conversion_webhook_secret_id" {
  description = "Secret ID for conversion webhook"
  value       = google_secret_manager_secret.conversion_webhook.secret_id
}
