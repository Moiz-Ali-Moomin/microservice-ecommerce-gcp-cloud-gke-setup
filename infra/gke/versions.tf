terraform {
  required_version = ">= 1.6"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }

    google-beta = {
      source  = "hashicorp/google-beta"
      version = ">= 5.0"
    }
  }

  # ----------------------------
  # Remote State (PRODUCTION)
  # ----------------------------
  # Uncomment for prod use
  #
  # backend "gcs" {
  #   bucket  = "tf-state-ecommerce-prod"
  #   prefix  = "gke/platform"
  #   project = "your-project-id"
  #   location = "us"
  # }
}
