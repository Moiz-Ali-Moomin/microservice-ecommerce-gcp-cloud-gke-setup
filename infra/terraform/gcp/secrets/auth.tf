resource "google_secret_manager_secret" "auth_jwt_secret" {
  secret_id = "auth-jwt-secret"
  labels = merge(local.common_labels, {
    service = "auth-service"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "auth_jwt_secret" {
  secret      = google_secret_manager_secret.auth_jwt_secret.id
  secret_data = var.auth_jwt_secret
}

resource "google_secret_manager_secret" "auth_db_user" {
  secret_id = "auth-db-user"
  labels = merge(local.common_labels, {
    service = "auth-service"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "auth_db_user" {
  secret      = google_secret_manager_secret.auth_db_user.id
  secret_data = var.auth_db_user
}
