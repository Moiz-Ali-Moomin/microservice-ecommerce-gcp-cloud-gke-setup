############################
# Global / Environment
############################

variable "project_id" {
  description = "The GCP Project ID"
  type        = string
  nullable    = false
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

############################
# Optional future-proofing
############################

variable "enable_deletion_protection" {
  description = "Enable deletion protection for critical resources"
  type        = bool
  default     = false
}
