# Grant External Secrets Operator access to ALL secrets managed by this stack
resource "google_project_iam_member" "secrets_reader_access" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${var.secrets_reader_sa_email}"
  
  # Note: This grants access to ALL secrets in the project.
  # This is consistent with a platform-level secrets infrastructure.
}
