############################
# Global / Environment
############################

variable "project_id" {
  description = "The GCP Project ID"
  type        = string
  nullable    = false

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{4,28}[a-z0-9]$", var.project_id))
    error_message = "project_id must be a valid GCP project ID (6-30 characters, lowercase, digits, hyphens)"
  }
}


variable "environment" {
  description = "Deployment environment (e.g. dev, staging, prod)"
  type        = string
  default     = "prod"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod"
  }
}

variable "region" {
  description = "The GCP Region"
  type        = string
  default     = "us-central1"

  validation {
    condition     = can(regex("^[a-z]+-[a-z]+[0-9]$", var.region))
    error_message = "region must be a valid GCP region code (e.g. us-central1)"
  }
}


############################
# GKE
############################

variable "cluster_name" {
  description = "The name of the GKE cluster"
  type        = string
  default     = "ecommerce-platform-prod"

  validation {
    condition     = length(var.cluster_name) > 3 && length(var.cluster_name) < 40
    error_message = "cluster_name must be between 4 and 39 characters"
  }
}

variable "master_ipv4_cidr_block" {
  description = "CIDR block for the GKE master network"
  type        = string
  default     = "172.16.0.0/28"

  validation {
    condition     = can(cidrnetmask(var.master_ipv4_cidr_block))
    error_message = "master_ipv4_cidr_block must be a valid CIDR block"
  }
}


############################
# Networking
############################

variable "network_name" {
  description = "The VPC network name"
  type        = string
  default     = "gke-vpc"
}

variable "subnet_name" {
  description = "The subnet name"
  type        = string
  default     = "gke-subnet"
}

variable "subnet_cidr" {
  description = "Primary CIDR range for the subnet"
  type        = string
  default     = "10.0.0.0/20"

  validation {
    condition     = can(cidrnetmask(var.subnet_cidr))
    error_message = "subnet_cidr must be a valid CIDR block"
  }
}

variable "pods_cidr_range" {
  description = "CIDR range for GKE pods"
  type        = string
  default     = "10.4.0.0/14"

  validation {
    condition     = can(cidrnetmask(var.pods_cidr_range))
    error_message = "pods_cidr_range must be a valid CIDR block"
  }
}

variable "services_cidr_range" {
  description = "CIDR range for GKE services"
  type        = string
  default     = "10.8.0.0/20"

  validation {
    condition     = can(cidrnetmask(var.services_cidr_range))
    error_message = "services_cidr_range must be a valid CIDR block"
  }
}


############################
# Node Pools
############################

variable "general_machine_type" {
  description = "Machine type for the general node pool"
  type        = string
  default     = "e2-standard-4"
}

variable "general_min_count" {
  description = "Minimum node count for general pool"
  type        = number
  default     = 1
}

variable "general_max_count" {
  description = "Maximum node count for general pool"
  type        = number
  default     = 5
}

variable "analytics_machine_type" {
  description = "Machine type for the analytics node pool"
  type        = string
  default     = "n2-highmem-8"
}

############################
# Artifact Registry
############################

variable "registry_id" {
  description = "ID for the Artifact Registry repository"
  type        = string
  default     = "ecommerce-repo"
}

############################
# Optional future-proofing
############################

variable "enable_deletion_protection" {
  description = "Enable deletion protection for critical resources"
  type        = bool
  default     = false
}
