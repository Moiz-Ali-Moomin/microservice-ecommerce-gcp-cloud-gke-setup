resource "google_storage_bucket" "terraform_state" {
  name     = "ecommerce-tf-state"
  location = var.region
  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      num_newer_versions = 10
    }
    action {
      type = "Delete"
    }
  }
}