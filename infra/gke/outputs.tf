output "cluster_name" {
  value = google_container_cluster.primary.name
}

output "cluster_endpoint" {
  value = google_container_cluster.primary.endpoint
}

output "cluster_ca_certificate" {
  sensitive = true
  value     = google_container_cluster.primary.master_auth[0].cluster_ca_certificate
}

output "network_name" {
  value = google_compute_network.gke_network.name
}

output "subnet_name" {
  value = google_compute_subnetwork.gke_subnet.name
}
