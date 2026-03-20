variable "project_id" {
  description = "The GCP project ID where all resources will be provisioned."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{4,28}[a-z0-9]$", var.project_id))
    error_message = "project_id must be a valid GCP project ID (lowercase letters, digits and hyphens, 6-30 chars)."
  }
}

variable "region" {
  description = "Primary GCP region for cluster deployment."
  type        = string
  default     = "us-central1"

  validation {
    condition     = can(regex("^[a-z]+-[a-z]+[0-9]$", var.region))
    error_message = "region must be a valid GCP region (e.g. us-central1, europe-west4)."
  }
}

variable "environment" {
  description = "Deployment environment label (dev | staging | prod)."
  type        = string
  default     = "prod"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}

variable "owner" {
  description = "Team or individual owner of this infrastructure (for billing attribution)."
  type        = string
  default     = "platform-engineering"
}

variable "cost_center" {
  description = "Cost center code for billing attribution."
  type        = string
  default     = "ecommerce-platform"
}

variable "cells" {
  description = "Map of cell definitions to deploy. Each cell is an isolated GKE cluster."
  type = map(object({
    cell_id           = number
    region            = optional(string) # Overrides var.region if set
    max_pods_per_node = optional(number, 110)
  }))
  default = {
    "cell-1" = { cell_id = 1 }
  }

  validation {
    condition     = alltrue([for k, v in var.cells : v.cell_id >= 1 && v.cell_id <= 99])
    error_message = "cell_id must be between 1 and 99 (used for /16 CIDR allocation)."
  }
}

variable "master_authorized_networks" {
  description = "List of CIDR blocks allowed to access the GKE control plane. Defaults to no public access."
  type = list(object({
    cidr_block   = string
    display_name = string
  }))
  default = []
}

variable "cmek_key_id" {
  description = "Cloud KMS key ID for CMEK encryption of the etcd database. Leave empty to use Google-managed keys."
  type        = string
  default     = ""
  sensitive   = true
}
