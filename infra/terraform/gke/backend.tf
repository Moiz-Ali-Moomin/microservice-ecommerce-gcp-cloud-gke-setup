# GCS remote backend with state locking via native GCS object versioning.
# Bucket must be created out-of-band (bootstrap/Makefile) before `terraform init`.
terraform {
  backend "gcs" {
    bucket = "ecommerce-tf-state"
    prefix = "gke/prod"
  }
}
