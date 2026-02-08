  resource "google_service_account" "secrets_reader" {
    project      = var.project_id
    account_id   = "secrets-reader"
    display_name = "Secrets Reader for External Secrets Operator"
  }

  # Grant the Service Account access to Secret Manager secrets
  resource "google_project_iam_member" "secrets_reader_access" {
    project = var.project_id
    role    = "roles/secretmanager.secretAccessor"
    member  = "serviceAccount:${google_service_account.secrets_reader.email}"
  }

  # Allow the Kubernetes Service Account to impersonate the GCP Service Account
  resource "google_service_account_iam_binding" "secrets_reader_wif" {
    service_account_id = google_service_account.secrets_reader.name
    role               = "roles/iam.workloadIdentityUser"

    members = [
      "serviceAccount:${var.project_id}.svc.id.goog[${var.k8s_namespace}/${var.k8s_sa_name}]"
    ]
  }
