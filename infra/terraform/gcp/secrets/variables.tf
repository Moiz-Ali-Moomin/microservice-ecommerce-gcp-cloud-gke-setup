variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "secrets_reader_sa_email" {
  description = "Email of the GCP Service Account used by External Secrets Operator (created in bootstrap-identity)"
  type        = string
}

# -----------------------------------------------------------------------------
# Sensitive Variables (Secret Values)
# -----------------------------------------------------------------------------

variable "postgres_app_password" {
  description = "Password for the Postgres application user"
  type        = string
  sensitive   = true
}

variable "redis_password" {
  description = "Password for Redis authentication"
  type        = string
  sensitive   = true
}

variable "kafka_password" {
  description = "SASL Password for Kafka authentication"
  type        = string
  sensitive   = true
}

variable "airflow_db_password" {
  description = "Password for Airflow metadata database"
  type        = string
  sensitive   = true
}

variable "metabase_db_password" {
  description = "Password for Metabase database"
  type        = string
  sensitive   = true
}
