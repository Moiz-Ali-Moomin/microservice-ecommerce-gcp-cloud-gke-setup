resource "google_monitoring_dashboard" "gke_dashboard" {
  project = var.project_id

  dashboard_json = <<EOF
{
  "displayName": "GKE Cluster Overview",
  "gridLayout": {
    "columns": 2,
    "widgets": [
      {
        "title": "CPU Usage (by Cluster / Namespace)",
        "xyChart": {
          "dataSets": [
            {
              "plotType": "LINE",
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "metric.type=\"kubernetes.io/container/cpu/core_usage_time\" resource.type=\"k8s_container\"",
                  "aggregation": {
                    "alignmentPeriod": "60s",
                    "perSeriesAligner": "ALIGN_RATE",
                    "groupByFields": [
                      "resource.label.cluster_name",
                      "resource.label.namespace_name"
                    ]
                  }
                }
              }
            }
          ],
          "yAxis": {
            "label": "CPU cores",
            "scale": "LINEAR"
          }
        }
      },
      {
        "title": "Memory Usage (by Cluster / Namespace)",
        "xyChart": {
          "dataSets": [
            {
              "plotType": "LINE",
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "metric.type=\"kubernetes.io/container/memory/used_bytes\" resource.type=\"k8s_container\"",
                  "aggregation": {
                    "alignmentPeriod": "60s",
                    "perSeriesAligner": "ALIGN_MEAN",
                    "groupByFields": [
                      "resource.label.cluster_name",
                      "resource.label.namespace_name"
                    ]
                  }
                }
              }
            }
          ],
          "yAxis": {
            "label": "Bytes",
            "scale": "LINEAR"
          }
        }
      }
    ]
  }
}
EOF
}
