variable "project_id" {
  type = string
}

variable "github_org" {
  type = string
}

variable "github_repo" {
  type = string
}

variable "workload_identity_pool_id" {
  type = string
}

variable "oidc_provider_id" {
  type = string
}

variable "service_account_name" {
  type    = string
  default = "terraform-sa"
}

variable "k8s_namespace" {
  description = "Kubernetes namespace for External Secrets Operator"
  type        = string
  default     = "ecommerce"
}

variable "k8s_sa_name" {
  description = "Kubernetes Service Account name for External Secrets Operator"
  type        = string
  default     = "external-secrets-sa"
}