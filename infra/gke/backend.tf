terraform {
  backend "gcs" {
    bucket = "ecommerce-tf-state"
    prefix = "platform/gcp/gke"
  }
}
