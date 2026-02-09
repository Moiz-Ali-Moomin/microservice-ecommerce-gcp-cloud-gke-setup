# --------------------------------------------------
# GKE Node Service Account
# --------------------------------------------------
# NOTE:
# - account_id must be 6–30 chars
# - lowercase letters, numbers, hyphens only
# - do NOT derive from cluster name (can exceed limit)
# --------------------------------------------------

resource "google_service_account" "gke_nodes" {
  account_id   = "gke-nodes-sa"
  display_name = "GKE Nodes Service Account"
}

# --------------------------------------------------
# Logging (required for GKE nodes)
# --------------------------------------------------

resource "google_project_iam_member" "gke_nodes_log_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}

# --------------------------------------------------
# Metrics (required for GKE / Cloud Monitoring)
# --------------------------------------------------

resource "google_project_iam_member" "gke_nodes_metric_writer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}

resource "google_project_iam_member" "gke_nodes_monitoring_viewer" {
  project = var.project_id
  role    = "roles/monitoring.viewer"
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}

# --------------------------------------------------
# Artifact Registry (pull container images)
# --------------------------------------------------

resource "google_project_iam_member" "gke_nodes_artifact_registry" {
  project = var.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}
