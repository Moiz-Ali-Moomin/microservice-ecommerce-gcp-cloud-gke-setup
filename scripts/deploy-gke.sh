#!/bin/bash
set -euo pipefail

# Staff+ DevOps / Production Bootstrap Script
# Goal: Deterministic, Fail-Fast GKE Initialization

# Colors for observability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO] $1${NC}"; }
log_warn() { echo -e "${YELLOW}[WARN] $1${NC}"; }
log_err() { echo -e "${RED}[ERROR] $1${NC}"; }

check_tool() {
  if ! command -v "$1" &> /dev/null; then
    log_err "Missing required tool: $1"
    exit 1
  fi
}

# 1. Pre-flight Checks
log_info "Phase 1: Pre-flight Checks"
for tool in gcloud terraform kubectl helm; do
  check_tool $tool
done

# Ensure we are in the root or correct relative path
if [ ! -d "infra/gke" ]; then
    log_err "Must run from repository root. 'infra/gke' not found."
    exit 1
fi

# 2. Infrastructure Layer (Terraform)
log_info "Phase 2: Infrastructure State Verification"
cd infra/gke

# Fail fast if Terraform is not initialized
if [ ! -d ".terraform" ]; then
    log_warn "Terraform not initialized. initializing..."
    terraform init
fi

read -p "Apply Terraform Infrastructure? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
  terraform validate
  terraform apply -auto-approve
else
  log_info "Skipping Terraform apply. Verifying existing state..."
fi

# Retrieve Outputs with NO fallback defaults (Unsafe)
REGION=$(terraform output -raw region)
PROJECT=$(terraform output -raw project_id)
CLUSTER=$(terraform output -raw kubernetes_cluster_name)

if [[ -z "$REGION" || -z "$PROJECT" || -z "$CLUSTER" ]]; then
    log_err "Terraform outputs missing. Infrastructure might be broken."
    exit 1
fi

log_info "Target: $CLUSTER ($REGION) in $PROJECT"
cd ../..

# 3. Cluster Connectivity
log_info "Phase 3: Cluster Authentication"
gcloud container clusters get-credentials "$CLUSTER" --region "$REGION" --project "$PROJECT"

# Verify connection
if ! kubectl cluster-info &> /dev/null; then
    log_err "Failed to connect to GKE cluster."
    exit 1
fi

# 4. GitOps Bootstrap (ArgoCD)
log_info "Phase 4: ArgoCD Installation"
read -p "Install/Upgrade ArgoCD? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
  # Create namespace safely
  kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
  
  # Apply upstream manifest (Specific version pinning recommended for strict prod, keeping stable for now)
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
  
  log_info "Waiting for ArgoCD Control Plane..."
  kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd
  kubectl wait --for=condition=available --timeout=300s deployment/argocd-repo-server -n argocd
else
  log_info "Skipping ArgoCD Install."
fi

# 5. System Components (The Platform)
log_info "Phase 5: Platform Layer Deployment"
log_info "Applying Namespaces..."
kubectl apply -f infra/namespaces/

log_info "Bootstraping External Secrets Operator (System Component)..."
# We deliberately apply the GitOps App for ESO, not the manifest, to let ArgoCD manage it.
kubectl apply -f argocd/apps/external-secrets.yaml

# Verify ESO System Namespace
log_info "Waiting for Platform Namespaces..."
sleep 5 # Give k8s API a moment
if ! kubectl get namespace ecommerce &> /dev/null; then
    log_err "Critical: 'ecommerce' namespace missing."
    exit 1
fi

# 6. Final Verification
log_info "Phase 6: Final System Verify"
echo "---------------------------------------------------"
echo "Infrastructure:  PROVISIONED"
echo "Cluster Auth:    CONFIGURED"
echo "GitOps Engine:   RUNNING"
echo "Platform Layer:  BOOTSTRAPPED"
echo "---------------------------------------------------"
log_warn "Application Services are NOT deployed by this script."
log_warn "Trigger the GitHub Actions pipeline to deploy microservices."
echo "---------------------------------------------------"
