# ============================================================================
# Cell Module — Isolated GKE Autopilot Cluster
#
# A "Cell" is a fully autonomous GKE cluster scaled to ~500k users.
# Configuration drift here NEVER affects other cells.
# ============================================================================

variable "project_id"        { type = string }
variable "region"            { type = string }
variable "cell_id"           { type = number }
variable "max_pods_per_node" { type = number; default = 110 }
variable "common_labels"     { type = map(string); default = {} }
variable "cmek_key_id"       { type = string; default = "" }
variable "master_authorized_networks" {
  type    = list(object({ cidr_block = string; display_name = string }))
  default = []
}

# ---------------------------------------------------------------------------
# Dedicated VPC Subnetwork — isolated IP space for this cell
# ---------------------------------------------------------------------------
resource "google_compute_subnetwork" "cell_subnet" {
  name          = "vpc-sub-cell-${var.cell_id}-${var.region}"
  ip_cidr_range = "10.${var.cell_id}.0.0/16"
  region        = var.region
  network       = "projects/${var.project_id}/global/networks/ecommerce-vpc"

  labels                   = var.common_labels
  private_ip_google_access = true

  secondary_ip_range {
    range_name    = "pod-ranges-${var.cell_id}"
    ip_cidr_range = "10.${var.cell_id + 100}.0.0/14"
  }
  secondary_ip_range {
    range_name    = "service-ranges-${var.cell_id}"
    ip_cidr_range = "10.${var.cell_id + 200}.0.0/20"
  }

  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }
}

# ---------------------------------------------------------------------------
# Regional Autopilot GKE Cluster
# ---------------------------------------------------------------------------
resource "google_container_cluster" "cell" {
  name     = "ecommerce-cell-${var.cell_id}"
  location = var.region

  enable_autopilot = true

  resource_labels = var.common_labels

  network    = google_compute_subnetwork.cell_subnet.network
  subnetwork = google_compute_subnetwork.cell_subnet.name

  ip_allocation_policy {
    cluster_secondary_range_name  = google_compute_subnetwork.cell_subnet.secondary_ip_range[0].range_name
    services_secondary_range_name = google_compute_subnetwork.cell_subnet.secondary_ip_range[1].range_name
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.${var.cell_id}.0/28"
  }

  # Restrict control plane access to approved CIDR blocks
  dynamic "master_authorized_networks_config" {
    for_each = length(var.master_authorized_networks) > 0 ? [1] : []
    content {
      dynamic "cidr_blocks" {
        for_each = var.master_authorized_networks
        content {
          cidr_block   = cidr_blocks.value.cidr_block
          display_name = cidr_blocks.value.display_name
        }
      }
    }
  }

  # Secretless GCP API access from pods via Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  # CMEK encryption of etcd — only provisioned when a key is provided
  dynamic "database_encryption" {
    for_each = var.cmek_key_id != "" ? [var.cmek_key_id] : []
    content {
      state    = "ENCRYPTED"
      key_name = database_encryption.value
    }
  }

  # Anthos Fleet registration for Hub GitOps management
  fleet {
    project = var.project_id
  }

  # STABLE channel: tested, enterprise-SLA releases
  release_channel {
    channel = "STABLE"
  }

  # Maintenance window: Tuesdays 03:00–05:00 UTC (lowest traffic window)
  maintenance_policy {
    recurring_window {
      start_time = "2024-01-01T03:00:00Z"
      end_time   = "2024-01-01T05:00:00Z"
      recurrence = "FREQ=WEEKLY;BYDAY=TU"
    }
  }

  # Binary Authorization enforces only Cosign-signed images (via Kyverno policy)
  binary_authorization {
    evaluation_mode = "PROJECT_SINGLETON_POLICY_ENFORCE"
  }

  # Route all logs to Cloud Logging for central analysis
  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  # Managed Prometheus for metrics (backs Grafana dashboards)
  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]
    managed_prometheus {
      enabled = true
    }
  }

  lifecycle {
    prevent_destroy = true
    ignore_changes  = [node_pool, initial_node_count]
  }
}

# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------
output "cluster_name" {
  description = "The name of the GKE cluster for this cell."
  value       = google_container_cluster.cell.name
}

output "cluster_endpoint" {
  description = "The control plane endpoint. Marked sensitive; used by CI/CD."
  value       = google_container_cluster.cell.endpoint
  sensitive   = true
}

output "subnet_name" {
  description = "The VPC subnet name provisioned for this cell."
  value       = google_compute_subnetwork.cell_subnet.name
}
