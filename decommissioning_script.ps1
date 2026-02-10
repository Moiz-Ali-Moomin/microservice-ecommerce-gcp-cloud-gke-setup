# decommissioning_script.ps1
# Systematically nuke all resources to save money

Write-Host "--- 🚀 Starting Full Platform Decommissioning ---" -ForegroundColor Cyan

# 1. Application and Platform Helm Releases (Corrected for namespaces)
Write-Host "Step 1: Uninstalling Helm Releases..." -ForegroundColor Yellow
$helmJson = helm list -A --output json | ConvertFrom-Json
foreach ($release in $helmJson) {
    Write-Host "Uninstalling $($release.name) from namespace $($release.namespace)..."
    helm uninstall $release.name -n $release.namespace --wait
}

# 2. Custom Kubernetes Resources (CRs)
Write-Host "Step 2: Deleting Custom Resources..." -ForegroundColor Yellow
kubectl delete sparkapplication --all -A --ignore-not-found
kubectl delete kafka --all -A --ignore-not-found
kubectl delete clustersecretstore --all --ignore-not-found

# 3. Manually Applied Deployments
Write-Host "Step 3: Deleting Manual Deployments..." -ForegroundColor Yellow
kubectl delete deployment metabase -n tooling-metabase --ignore-not-found
kubectl delete -f https://strimzi.io/install/latest?namespace=kafka --ignore-not-found

# 4. Namespaces (Cleanup)
Write-Host "Step 4: Deleting Custom Namespaces..." -ForegroundColor Yellow
$namespaces = "apps-async", "apps-core", "apps-public", "data-postgres", "data-redis", "external-secrets", "kafka", "platform-observability", "services", "spark-operator", "tooling", "tooling-airflow", "tooling-metabase", "tooling-sonarqube"
foreach ($ns in $namespaces) {
    Write-Host "Deleting namespace $ns..."
    kubectl delete namespace $ns --ignore-not-found --wait=false
}

Write-Host "--- ✅ Kubernetes Surface Area Cleaned ---" -ForegroundColor Green
Write-Host "Next Step: Run Terraform Destroy in reversed order." -ForegroundColor Cyan
