resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  location = var.region
  project  = var.project_id

  # We can't create a cluster with no node pool defined,
  # so we create the smallest possible default node pool
  # and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1

  network    = google_compute_network.gke_network.id
  subnetwork = google_compute_subnetwork.gke_subnet.id

  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"

  # VPC-native networking
  ip_allocation_policy {
    cluster_secondary_range_name  = "gke-pods"
    services_secondary_range_name = "gke-services"
  }

  network_policy {
    enabled = true
  }

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false # true = fully private (VPN/Bastion needed)
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }

  # Dataplane V2 (eBPF)
  datapath_provider = "ADVANCED_DATAPATH"

  # Maintenance Window
  maintenance_policy {
    daily_maintenance_window {
      start_time = "03:00"
    }
  }

  release_channel {
    channel = "REGULAR"
  }

  # Cost visibility
  cost_management_config {
    enabled = true
  }

  deletion_protection = false # set true in locked-down prod
}
