# 🚀 E-Commerce Platform Deployment Guide

This document outlines the proven steps to deploy the entire e-commerce infrastructure and microservices stack to Google Kubernetes Engine (GKE). It incorporates all fixes for known issues.

## 🛠️ Prerequisites
- **Tools**: `gcloud`, `terraform`, `kubectl`, `helm`, `istioctl`
- **Access**: GCP Project Owner/Editor role

---

## 1. Infrastructure Provisioning (Terraform)
Navigate to `infra/terraform/gcp` and apply the modules in order.

### 1.1 Bootstrap Identity & Backend
Sets up Workload Identity Federation and GCS backend.
```bash
cd infra/terraform/gcp/bootstrap-identity
terraform init
terraform apply -var="project_id=<YOUR_PROJECT_ID>"
```
> **✅ Fix Applied**: Default variables for `k8s_namespace` and `k8s_sa_name` now point to `external-secrets`, ensuring correct IAM bindings are created automatically.

### 1.2 Secrets Manager
Provisions Google Secret Manager.
```bash
cd ../secrets
terraform apply
```

### 1.3 GKE Cluster
Provisions the Kubernetes cluster.
```bash
cd ../gke
terraform apply
```

---

## 2. Base Platform Services

### 2.1 Connect to Cluster
```bash
gcloud container clusters get-credentials ecommerce-cluster --region <REGION>
```

### 2.2 External Secrets Operator
Syncs GCP Secrets to Kubernetes.
```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
    -n external-secrets --create-namespace \
    --set installCRDs=true
```

#### Apply ClusterSecretStore
```bash
kubectl apply -f infra/external-secrets/cluster-secret-store.yaml
```
> **✅ Fix Applied**: Used a static manifest with `v1` API version and correct Project ID to bypass broken Helm templates.

### 2.3 Kafka (Strimzi Operator)
Deploys Kafka in KRaft mode.
```bash
# Install Operator
kubectl create namespace kafka
kubectl apply -f infra/kafka/operator/strimzi-install.yaml

# Deploy Cluster
kubectl apply -f infra/kafka/cluster/kafka.yaml
```
> **✅ Fix Applied**: Used official Strimzi manifests to resolve ClusterRoleBinding conflicts.

---

## 3. Databases

### 3.1 PostgreSQL (Bitnami)
```bash
helm install postgres bitnami/postgresql \
    -n data-postgres --create-namespace \
    -f infra/databases/postgres/values.yaml \
    --set metrics.serviceMonitor.enabled=false
```
> **✅ Fix Applied**: Disabled `serviceMonitor` to prevent failure when Prometheus Operator is not installed.

### 3.2 Redis (Bitnami)
```bash
helm install redis bitnami/redis \
    -n data-redis --create-namespace \
    -f infra/databases/redis/values.yaml \
    --set metrics.serviceMonitor.enabled=false
```

---

## 4. Microservices Deployment

### 4.1 Sync Secrets
Ensure all `ExternalSecret` manifests are applied first.
```bash
kubectl apply -f infra/secrets/
kubectl apply -f infra/app-secrets/
```
> **✅ Fix Applied**: All manifests updated to `apiVersion: external-secrets.io/v1`.

### 4.2 Deploy Apps
Deploy all services using Helm or Manifests.
```bash
# Example for User Service
helm upgrade --install user-service services/user-service/helm \
    -n apps-core --create-namespace \
    -f services/user-service/helm/values.yaml

# Repeat for all services in services/ directory...
```
> **Key Services**: `auth-service`, `user-service`, `product-service`, `cart-service`, `order-service` (and others).

---

## 5. Service Mesh (Istio)

### 5.1 Control Plane
Install minimal profile to avoid conflicts.
```bash
istioctl install --set profile=minimal -y
```

### 5.2 Ingress Gateway
Deploy custom gateway.
```bash
kubectl apply -f infra/istio/gateway-workload/
```
> **✅ Fixes Applied**:
> - Updated image to `istio/proxyv2:1.25.2` to match control plane.
> - Added missing volumes for `/var/run/secrets/workload-spiffe-uds`.
> - Fixed YAML syntax errors.

### 5.3 Traffic Routing
Configure Gateway and VirtualServices.
```bash
kubectl apply -f infra/istio/ingress/
kubectl apply -f infra/istio/ingress/virtual-services/
```
> **✅ Fix Applied**: Updated `ecommerce-all.yaml` to point routes to correct namespaces (`tooling`, `ecommerce`, `apps-async`).

---

## 7. Observability Stack

### 7.1 Install Helm Repos
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

### 7.2 Deploy Components
Deploy to `platform-observability` namespace.

**Prometheus** (Metrics):
```bash
helm upgrade --install prometheus prometheus-community/prometheus \
    -n platform-observability --create-namespace \
    -f infra/observability/prometheus/values.yaml
```

**Loki** (Logs):
```bash
helm upgrade --install loki grafana/loki \
    -n platform-observability \
    -f infra/observability/loki/values.yaml
```
> **✅ Fix Applied**: Updated schema config to `tsdb/v13` and disabled scalable replicas in `values.yaml` to fix crash loop.

**Tempo** (Tracing):
```bash
helm upgrade --install tempo grafana/tempo \
    -n platform-observability \
    -f infra/observability/tempo/values.yaml
```

**Grafana** (Visualization):
```bash
helm upgrade --install grafana grafana/grafana \
    -n platform-observability \
    -f infra/observability/grafana/values.yaml
```

---

## 8. Verification

### Check Pod Status
All pods should be `1/1 Running`.
```bash
kubectl get pods -A
```

### Access Public Endpoints
Get the LoadBalancer IP:
```bash
kubectl get svc -n istio-system istio-ingressgateway
```
Test routes:
- `http://<IP>/` -> Storefront
- `http://<IP>/api-gateway/health` -> API Gateway
- `http://<IP>/product` -> Product Service

---

## 9. Current Service Issues (Active Troubleshooting)

The following services are currently under active remediation:

### 9.1 Metabase (`tooling-metabase`)
*   **Status**: `Stabilizing / Initializing`
*   **Root Cause**: Mismatched taints (`workload=analytics`) and aggressive liveness probes during long migrations.
*   **Fix**: Applied SRE-hardened manifest with `startupProbes`, `initContainers`, and proper `tolerations`.
*   **Action**: Wait ~10 minutes for Clojure startup and DB migrations to complete.

### 9.2 Spark Job (`apps-async`)
*   **Status**: `Stuck / Pending`
*   **Root Cause**: Spark Operator is restricted to the `default` namespace and doesn't watch `apps-async`.
*   **Fix**: Patching operator deployment to remove `--namespaces=default` argument and enable global watching.
*   **Action**: Re-trigger `SparkApplication` after operator restart.

---

## 9. Current Service Issues (Active Troubleshooting)

The following services are currently under active remediation:

### 9.1 Metabase (`tooling-metabase`)
*   **Status**: `Stabilizing / Initializing`
*   **Root Cause**: Mismatched taints (`workload=analytics`) and aggressive liveness probes during long migrations.
*   **Fix**: Applied SRE-hardened manifest with `startupProbes`, `initContainers`, and proper `tolerations`.
*   **Action**: Wait ~10 minutes for Clojure startup and DB migrations to complete.

### 9.2 Spark Job (`apps-async`)
*   **Status**: `Stuck / Pending`
*   **Root Cause**: Spark Operator is restricted to the `default` namespace and doesn't watch `apps-async`.
*   **Fix**: Patching operator deployment to remove `--namespaces=default` argument and enable global watching.
*   **Action**: Re-trigger `SparkApplication` after operator restart.

---

## 🛑 Troubleshooting Summary
If you encounter issues, check `deployment-troubleshooting.md` in the `brain/` directory for detailed root cause analysis of:
- Helper chart values vs Manifests
- IAM Permission Denied errors
- Istio CrashLoopBackOffs
- Service-specific configuration mismatches
- Service-specific configuration mismatches
