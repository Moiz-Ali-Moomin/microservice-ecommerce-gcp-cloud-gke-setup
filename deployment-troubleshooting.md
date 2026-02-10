# Deployment Troubleshooting & Fixes

## 1. Database Deployment (PostgreSQL & Redis)

### 🔴 The Issue
Attempted to run `kubectl apply -f infra/databases/postgres/values.yaml`, which failed with:
```
error: error validating data: [apiVersion not set, kind not set]
```

### 🔍 Root Cause
The files were **Helm Chart Values**, not Kubernetes manifests. They are configuration files designed to be used with `helm install`, not `kubectl apply`.

### ✅ The Fix
Used Helm with the Bitnami charts:
```bash
# Correct command
helm install postgres bitnami/postgresql -f infra/databases/postgres/values.yaml
```

---

## 2. ServiceMonitor CRD Missing

### 🔴 The Issue
PostgreSQL installation failed with:
```
no matches for kind "ServiceMonitor" in version "monitoring.coreos.com/v1"
```

### 🔍 Root Cause
The Helm chart defaults to enabling metrics (`metrics.enabled: true`), which tries to create a `ServiceMonitor`. This resource requires the **Prometheus Operator** CRDs, which were not yet installed.

### ✅ The Fix
Disabled ServiceMonitor creation during install:
```bash
--set metrics.serviceMonitor.enabled=false
```

---

## 3. External Secrets API Version Mismatch

### 🔴 The Issue
Applying `ClusterSecretStore` and `ExternalSecret` manifests failed with:
```
no matches for kind "ClusterSecretStore" ... ensure CRDs are installed first
```

### 🔍 Root Cause
The installed External Secrets Operator was v0.9.x+ which uses `external-secrets.io/v1`.
The manifests were outdated, using the deprecated `external-secrets.io/v1beta1`.

### ✅ The Fix
Updated `apiVersion` in all manifests:
```yaml
- apiVersion: external-secrets.io/v1beta1
+ apiVersion: external-secrets.io/v1
```

---

## 4. IAM Permission Denied (Workload Identity)

### 🔴 The Issue
External Secrets failed to sync with `SecretSyncedError`:
```
IAM_PERMISSION_DENIED: iam.serviceAccounts.getAccessToken
```

### 🔍 Root Cause
Terraform defaults pointed Workload Identity to:
`serviceAccount:...[ecommerce/external-secrets-sa]`

But the actual deployment used:
`serviceAccount:...[external-secrets/external-secrets]`

The Namespace (`external-secrets` vs `ecommerce`) and SA Name (`external-secrets` vs `external-secrets-sa`) were mismatched.

### ✅ The Fix
Manually added the correct IAM binding:
```bash
gcloud iam service-accounts add-iam-policy-binding secrets-reader@... \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:PROJECT_ID.svc.id.goog[external-secrets/external-secrets]"
```

---

## 5. User Service Project ID

### 🔴 The Issue
`user-service` Helm values contained a placeholder/wrong Project ID:
`user-service@ecommerce-project.iam.gserviceaccount.com`

### ✅ The Fix
Updated `values.yaml` to the correct ID:
`user-service@ecommerce-microservice-53.iam.gserviceaccount.com`

---

## 6. Istio Ingress Gateway Deployment

### 🔴 The Issue (CrashLoopBackOff)
The custom Ingress Gateway pods failed to start with `CrashLoopBackOff`.

### 🔍 Root Causes
1.  **Version Mismatch**: `istiod` control plane was version `1.25.2`, but the gateway deployment image was pinned to `istio/proxyv2:1.22.1`.
2.  **Missing SDS Socket Volume**: The proxy failed to start with `connect error: No such file or directory` for `/var/run/secrets/workload-spiffe-uds/socket`.
3.  **YAML Syntax Error**: A duplicate `secretName` key prevented the corrected manifest from applying.

### ✅ The Fixes
1.  Updated image to `istio/proxyv2:1.25.2`.
2.  Added `emptyDir` volumes for `workload-socket` and `istio-proxy`.
3.  Fixed YAML syntax error.

---

## 7. Loki (Observability) CrashLoopBackOff

### 🔴 The Issue
Loki pods failed to start with "MULTIPLE CONFIG ERRORS" or deployment mode conflicts.

### 🔍 Root Causes
1.  **Schema Config**: `boltdb-shipper` (deprecated in v3.x) was used.
2.  **Deployment Mode**: Scalable targets were enabled by default in the chart, conflicting with `SingleBinary` mode.

### ✅ The Fixes
1.  Updated schema to `tsdb` version `v13`.
2.  Explicitly set `read.replicas: 0`, `write.replicas: 0`, `backend.replicas: 0` in `values.yaml`.
3.  Set `deploymentMode: SingleBinary`.

---

## 8. Async Services & Storefront CrashLoopBackOff

### 🔴 The Issue
Pods (`analytics-ingest`, `attribution`, `audit`, `notification`, `storefront`) were restart-looping or failing to be created.
1.  **CrashLoopBackOff**: Logs showed `Liveness probe failed: HTTP probe failed with statuscode: 404`.
2.  **Creation Failure**: Deployment controller showed `violated PodSecurity "baseline:latest": non-default capabilities (container "istio-init" must not include "NET_ADMIN")`.

### 🔍 Root Causes
1.  **Missing Health Endpoint**: The services' Go code did not expose `/health` (used by default in Helm charts), causing liveness probes to fail.
2.  **Pod Security Policy**: The namespaces `apps-async` and `apps-public` enforced `baseline` security, which blocks `NET_ADMIN` capability required by Istio's init container for iptables setup.

### ✅ The Fixes
1.  **Updated Probes**: Modified `deployment.yaml` for affected services to use `tcpSocket` on port `http` instead of `httpGet` on `/health`.
2.  **Updated Namespace Security**: Changed `pod-security.kubernetes.io/enforce` label from `baseline` to `privileged` for `apps-async` and `apps-public` namespaces to allow Istio injection.

---

## 9. Airflow API Server CrashLoopBackOff

### 🔴 The Issue
Airflow API Server pod crashed with:
```
invalid choice: 'api-server' (choose from 'webserver', 'scheduler', ...)
```

### 🔍 Root Cause
**Version mismatch between Helm chart and Docker image**:
- Helm Chart `apache-airflow/airflow` version `1.18.0` expects **Airflow 3.x** which has an `api-server` command
- Override with `3.0.2` image to match.

### ✅ Permanent Fix
Update `infra/airflow/values.yaml` to use `apache/airflow:3.0.2`.

---

## 10. SonarQube Pod Security Violation

### 🔴 The Issue
SonarQube StatefulSet failed to create pods due to `privileged` init container.

### 🔍 Root Cause
Namespace `tooling-sonarqube` had `baseline` security.

### ✅ Permanent Fix
Set `pod-security.kubernetes.io/enforce: privileged` on the namespace.
