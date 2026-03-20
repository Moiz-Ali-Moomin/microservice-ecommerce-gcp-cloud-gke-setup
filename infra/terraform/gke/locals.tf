locals {
  # Standard resource labels applied to every GCP resource for cost attribution and ops tooling.
  common_labels = {
    env         = var.environment
    owner       = var.owner
    cost-center = var.cost_center
    managed-by  = "terraform"
    repo        = "microservice-ecommerce-gcp-cloud-gke-setup"
  }

  # Resolved cells map: fills in optional region from the global var.
  resolved_cells = {
    for k, v in var.cells : k => merge(v, {
      region = coalesce(v.region, var.region)
    })
  }

  # Maintenance window: every Tuesday 03:00–05:00 UTC (low-traffic window)
  maintenance_recurrence = "FREQ=WEEKLY;BYDAY=TU"
  maintenance_start      = "2024-01-01T03:00:00Z"
  maintenance_end        = "2024-01-01T05:00:00Z"
}
