resource "google_compute_network" "gke_network" {
  name                    = var.network_name
  auto_create_subnetworks = false

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_compute_subnetwork" "gke_subnet" {
  name                     = var.subnet_name
  region                   = var.region
  network                  = google_compute_network.gke_network.id
  ip_cidr_range            = var.subnet_cidr
  private_ip_google_access = true

  secondary_ip_range {
    range_name    = "gke-pods"
    ip_cidr_range = var.pods_cidr_range
  }

  secondary_ip_range {
    range_name    = "gke-services"
    ip_cidr_range = var.services_cidr_range
  }

  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }
  lifecycle {
    prevent_destroy = true
  }
}
