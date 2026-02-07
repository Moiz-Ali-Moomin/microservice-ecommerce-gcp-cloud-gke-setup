provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  alias   = "beta"
  project = var.project_id
  region  = var.region
}

resource "google_container_cluster" "primary" {
  provider = google

  # stable features
}
