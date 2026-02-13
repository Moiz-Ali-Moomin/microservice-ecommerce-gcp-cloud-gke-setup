# JWT Secret for auth-service
resource "google_secret_manager_secret" "auth_service_jwt" {
  secret_id = "auth-service-jwt"   # ← FIXED NAME

  labels = merge(local.common_labels, {
    service = "auth-service"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "auth_service_jwt" {
  secret      = google_secret_manager_secret.auth_service_jwt.id
  secret_data = var.auth_jwt_secret
}


# Database secret for auth-service
resource "google_secret_manager_secret" "auth_service_db" {
  secret_id = "auth-service-db"   # ← FIXED NAME

  labels = merge(local.common_labels, {
    service = "auth-service"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "auth_service_db" {
  secret      = google_secret_manager_secret.auth_service_db.id
  secret_data = var.postgres_app_password
}


# Redis secret for auth-service
resource "google_secret_manager_secret" "auth_service_redis" {
  secret_id = "auth-service-redis"   # ← FIXED NAME

  labels = merge(local.common_labels, {
    service = "auth-service"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "auth_service_redis" {
  secret      = google_secret_manager_secret.auth_service_redis.id
  secret_data = var.redis_password
}
