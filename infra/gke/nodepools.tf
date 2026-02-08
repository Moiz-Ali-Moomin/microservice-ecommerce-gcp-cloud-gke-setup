resource "google_container_node_pool" "general" {
  name     = "general-pool"
  location = var.region
  cluster  = google_container_cluster.primary.name
  project  = var.project_id

  node_locations = ["${var.region}-a", "${var.region}-b"]

  node_count = 1

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  autoscaling {
    min_node_count = var.general_min_count
    max_node_count = var.general_max_count
  }

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  node_config {
    machine_type = var.general_machine_type
    preemptible  = false

    disk_size_gb = 100
    disk_type    = "pd-standard"

    service_account = google_service_account.gke_nodes.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }

    tags = ["gke-general"]

    labels = {
      env  = "prod"
      team = "platform"
      pool = "general"
    }
  }
}

resource "google_container_node_pool" "analytics" {
  name     = "analytics-pool"
  location = var.region
  cluster  = google_container_cluster.primary.name
  project  = var.project_id

  initial_node_count = 0

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  autoscaling {
    min_node_count = 0
    max_node_count = 3
  }

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  node_config {
    machine_type = var.analytics_machine_type

    # Spot instances for cost optimization.
    # Ensure workloads on this pool are tolerant of interruptions.
    spot = true

    disk_size_gb = 200
    disk_type    = "pd-ssd"

    service_account = google_service_account.gke_nodes.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    # Taint to ensure only dedicated analytics workloads run here.
    # Pods must tolerate this taint to be scheduled.
    taint {
      key    = "workload"
      value  = "analytics"
      effect = "NO_SCHEDULE"
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }

    tags = ["gke-analytics"]

    labels = {
      env      = "prod"
      team     = "data"
      workload = "analytics"
      pool     = "analytics"
    }
  }
}
