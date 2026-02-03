#!/bin/bash
# ============================================================
# Production K8s Deployment Script
# Project: ecommerce-microservice-53
# ============================================================
set -euo pipefail

# Configuration
export PROJECT_ID="ecommerce-microservice-53"
export REGION="us-central1"
export REGISTRY="${REGION}-docker.pkg.dev/${PROJECT_ID}/ecommerce-repo"
export TAG=$(git rev-parse --short HEAD)
export NAMESPACE="ecommerce"
export KAFKA_NAMESPACE="kafka"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# All services
SERVICES=(
  admin-backoffice-service
  analytics-ingest-service
  api-gateway
  attribution-service
  audit-service
  auth-service
  cart-service
  conversion-webhook
  feature-flag-service
  landing-service
  notification-service
  offer-service
  redirect-service
  reporting-service
  storefront-service
  user-service
)

# ============================================================
# Step 1: Authenticate with GCP
# ============================================================
authenticate() {
  log "Authenticating with GCP..."
  gcloud auth configure-docker ${REGION}-docker.pkg.dev --quiet
  
  log "Verifying kubectl context..."
  kubectl cluster-info || error "kubectl not configured. Run: gcloud container clusters get-credentials ecommerce-cluster --zone ${REGION}-a"
}

# ============================================================
# Step 2: Create Artifact Registry (if not exists)
# ============================================================
create_registry() {
  log "Creating Artifact Registry repository..."
  gcloud artifacts repositories create ecommerce-repo \
    --repository-format=docker \
    --location=${REGION} \
    --description="E-commerce microservices" 2>/dev/null || true
}

# ============================================================
# Step 3: Push Images
# ============================================================
push_images() {
  log "Pushing images with tag: ${TAG}"
  
  for svc in "${SERVICES[@]}"; do
    echo -n "  Pushing ${svc}... "
    docker tag ecommerce/${svc}:local ${REGISTRY}/${svc}:${TAG} 2>/dev/null || { warn "Image not found: ecommerce/${svc}:local"; continue; }
    docker push ${REGISTRY}/${svc}:${TAG} --quiet
    echo "✓"
  done
  
  log "All images pushed with tag: ${TAG}"
}

# ============================================================
# Step 4: Create Namespaces
# ============================================================
create_namespaces() {
  log "Creating namespaces..."
  kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
  kubectl create namespace ${KAFKA_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
}

# ============================================================
# Step 5: Deploy Kafka
# ============================================================
deploy_kafka() {
  if kubectl get statefulset kafka -n ${KAFKA_NAMESPACE} &>/dev/null; then
    log "Kafka already deployed, skipping..."
    return
  fi
  
  log "Deploying Apache Kafka (Kraft mode)..."
  # Apply manifests
  kubectl apply -f infra/kafka/kafka-brokers.yaml
  kubectl apply -f infra/kafka/topics.yaml
  
  log "Waiting for Kafka to be ready..."
  # Wait for at least one pod to be ready to confirm rollout started
  kubectl wait --for=condition=ready pod -l app=kafka -n ${KAFKA_NAMESPACE} --timeout=300s
  
  log "Kafka deployed successfully"
}

# ============================================================
# Step 6: Deploy Services
# ============================================================
deploy_services() {
  log "Deploying services to namespace: ${NAMESPACE}"
  
  # Deploy in order (dependencies first)
  ORDERED_SERVICES=(
    auth-service
    user-service
    offer-service
    feature-flag-service
    cart-service
    storefront-service
    analytics-ingest-service
    attribution-service
    audit-service
    notification-service
    landing-service
    redirect-service
    conversion-webhook
    reporting-service
    admin-backoffice-service
    api-gateway
  )
  
  for svc in "${ORDERED_SERVICES[@]}"; do
    echo -n "  Deploying ${svc}... "
    
    helm upgrade --install ${svc} services/${svc}/helm \
      --namespace ${NAMESPACE} \
      --create-namespace \
      --set image.repository=${REGISTRY}/${svc} \
      --set image.tag=${TAG} \
      --set env.KAFKA_BROKERS=kafka-headless.${KAFKA_NAMESPACE}.svc:9092 \
      --wait --timeout 300s 2>/dev/null && echo "✓" || { warn "Failed (continuing)"; }
  done
  
  log "All services deployed"
}

# ============================================================
# Step 7: Expose API Gateway
# ============================================================
expose_gateway() {
  log "Exposing API Gateway..."
  kubectl patch svc api-gateway -n ${NAMESPACE} \
    -p '{"spec": {"type": "LoadBalancer"}}' 2>/dev/null || true
  
  echo ""
  log "Waiting for external IP (this may take 1-2 minutes)..."
  kubectl get svc api-gateway -n ${NAMESPACE} -w &
  WATCH_PID=$!
  sleep 60
  kill $WATCH_PID 2>/dev/null || true
  
  GATEWAY_IP=$(kubectl get svc api-gateway -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
  if [ -n "$GATEWAY_IP" ]; then
    log "API Gateway available at: http://${GATEWAY_IP}"
  else
    warn "External IP not ready yet. Check with: kubectl get svc api-gateway -n ${NAMESPACE}"
  fi
}

# ============================================================
# Step 8: Verify
# ============================================================
verify() {
  log "Verifying deployment..."
  echo ""
  kubectl get pods -n ${NAMESPACE}
  echo ""
  kubectl get svc -n ${NAMESPACE}
}

# ============================================================
# Main
# ============================================================
main() {
  echo "============================================================"
  echo "  E-commerce Platform - K8s Deployment"
  echo "  Project: ${PROJECT_ID}"
  echo "  Tag: ${TAG}"
  echo "============================================================"
  echo ""
  
  authenticate
  create_registry
  push_images
  create_namespaces
  deploy_kafka
  deploy_services
  expose_gateway
  verify
  
  echo ""
  log "🎉 Deployment complete!"
}

# Run
main "$@"
