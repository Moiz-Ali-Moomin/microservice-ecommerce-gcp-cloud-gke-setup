output "cluster_names" {
  description = "Map of cell key → GKE cluster name."
  value       = { for k, m in module.cells : k => m.cluster_name }
}

output "cluster_endpoints" {
  description = "Map of cell key → GKE control plane endpoint. Marked sensitive."
  value       = { for k, m in module.cells : k => m.cluster_endpoint }
  sensitive   = true
}

output "workload_identity_pool" {
  description = "Workload Identity pool used by all cells for secretless GCP authentication."
  value       = "${var.project_id}.svc.id.goog"
}

output "artifact_registry_url" {
  description = "Container image registry URL."
  value       = "us-central1-docker.pkg.dev/${var.project_id}/ecommerce-repo"
}
