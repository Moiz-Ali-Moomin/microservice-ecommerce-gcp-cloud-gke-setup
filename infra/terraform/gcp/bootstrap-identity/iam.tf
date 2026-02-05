resource "google_service_account" "terraform" {
  project      = var.project_id
  account_id   = var.service_account_name
  display_name = "Terraform GitHub OIDC Runner"
}

resource "google_service_account_iam_binding" "wif" {
  service_account_id = google_service_account.terraform.name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "principalSet://iam.googleapis.com/projects/${data.google_project.this.number}/locations/global/workloadIdentityPools/${google_iam_workload_identity_pool.github.workload_identity_pool_id}/attribute.repository/${var.github_org}/${var.github_repo}"
  ]
}
