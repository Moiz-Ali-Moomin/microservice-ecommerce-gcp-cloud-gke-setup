<div align="center">

# 🛒 Microservice E-Commerce Platform

**Enterprise-grade, GKE-native, globally scalable e-commerce infrastructure**

[![CI/CD](https://github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/actions/workflows/ci.yml/badge.svg)](https://github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Go Version](https://img.shields.io/badge/Go-1.24-blue)
![Terraform](https://img.shields.io/badge/Terraform-1.6+-purple)

</div>

---

## 📐 Architecture

```
╔══════════════════════════════════════════════════════════════════════════╗
║                         INTERNET TRAFFIC                                ║
╚══════════════════════════════════════════════════════════════════════════╝
                                   │
                    ┌──────────────▼──────────────┐
                    │     Istio Ingress Gateway    │  TLS termination
                    │     (GKE Load Balancer)      │  WAF + DDoS
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │         API Gateway          │  Rate limiting
                    │      (Go / :8080)            │  Auth routing
                    └──┬──┬──┬──┬──┬──┬──┬───────┘  Request ID
                       │  │  │  │  │  │  │
          ┌────────────┘  │  │  │  │  │  └───────────────────┐
          │     ┌─────────┘  │  │  │  └──────────┐           │
          ▼     ▼            ▼  │  ▼             ▼           ▼
    ┌─────────┐ ┌──────────┐   │  ┌──────────┐ ┌──────────┐ ┌──────────┐
    │  Auth   │ │  Order   │   │  │  Cart    │ │  User    │ │Storefront│
    │ Service │ │ Service  │   │  │ Service  │ │ Service  │ │   Web    │
    └────┬────┘ └────┬─────┘   │  └────┬─────┘ └────┬─────┘ └──────────┘
         │           │         │       │             │
         └─────┬─────┘    ┌────▼────┐  └──────┬──────┘
               │          │ Product │         │
               │          │ Service │         │
               │          └─────────┘         │
               ▼                              ▼
    ┌───────────────────────────────────────────────┐
    │                 DATA LAYER                    │
    │  ┌──────────────┐  ┌──────┐  ┌────────────┐  │
    │  │  PostgreSQL  │  │Redis │  │   Kafka    │  │
    │  │  (Primary +  │  │Cache │  │  (Events)  │  │
    │  │   Replicas)  │  │      │  │            │  │
    │  └──────────────┘  └──────┘  └────────────┘  │
    └───────────────────────────────────────────────┘
               │
               ▼
    ┌───────────────────────────────────────────────┐
    │            OBSERVABILITY PLANE                │
    │  Prometheus → Grafana  |  Loki  |  Tempo      │
    │  OTel Collector → Distributed Tracing         │
    │  Alertmanager → PagerDuty / Slack             │
    └───────────────────────────────────────────────┘
```

### Cell-Based Scaling

```
┌────────────────────────────────────────────────────────────────┐
│                    GCP Project (Hub)                           │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│   │   Cell 1    │  │   Cell 2    │  │   Cell N    │          │
│   │ GKE Region  │  │ GKE Region  │  │ GKE Region  │          │
│   │ ~500k users │  │ ~500k users │  │ ~500k users │          │
│   │  us-central │  │ europe-west │  │  asia-east  │          │
│   └─────────────┘  └─────────────┘  └─────────────┘          │
│   ArgoCD Hub  |  Anthos Fleet  |  Terraform Cells             │
└────────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Tech Stack

| Layer | Technology | Version | Purpose |
|---|---|---|---|
| **Language** | Go | 1.24 | All microservices |
| **Container Runtime** | Docker BuildKit + Distroless | latest | Secure, minimal images |
| **Orchestration** | GKE Autopilot | STABLE channel | Multi-cell K8s clusters |
| **Service Mesh** | Istio | 1.20+ | mTLS, traffic management |
| **GitOps** | ArgoCD | 2.10+ | App-of-Apps pattern |
| **Infrastructure** | Terraform | ≥1.6 | GCS remote state |
| **Messaging** | Apache Kafka | 3.x | Event streaming |
| **Cache** | Redis | 7.x | Session + cache |
| **Database** | PostgreSQL | 15+ | Primary data store |
| **Metrics** | Prometheus + Grafana | latest | Dashboards + alerting |
| **Logging** | Loki + Grafana | latest | Centralized log aggregation |
| **Tracing** | OpenTelemetry + Tempo | latest | Distributed trace storage |
| **Secrets** | External Secrets Operator | 0.9+ | GCP Secret Manager sync |
| **Policy** | Kyverno | 1.12+ | Image signing enforcement |
| **CI/CD** | GitHub Actions | - | Multi-env pipeline |
| **Code Quality** | SonarQube + revive | - | Static analysis |
| **Security Scans** | Trivy | 0.16+ | Container + FS scanning |
| **Image Signing** | Cosign | 2.2+ | Keyless OIDC signing |

---

## 🚀 Microservices

| Service | Port | Responsibilities |
|---|---|---|
| `api-gateway` | 8080 | Routing, rate limiting, request ID injection |
| `auth-service` | 8080 | JWT issuance, login, signup, token validation |
| `user-service` | 8080 | User profiles, preferences |
| `order-service` | 8080 | Order creation with idempotency, Kafka emit |
| `cart-service` | 8080 | Shopping cart management |
| `product-service` | 8080 | Product catalog, inventory |
| `offer-service` | 8080 | Promotions, discount engine |
| `storefront-service` | 3000 | BFF for the main web app |
| `storefront-web` | 3000 | React/Next.js frontend |
| `landing-service` | 8080 | Campaign landing pages |
| `redirect-service` | 8080 | Click-tracking redirect |
| `notification-service` | 8080 | Email/SMS/push notifications |
| `reporting-service` | 8080 | Analytics + business reports |
| `analytics-ingest-service` | 8080 | Event ingestion pipeline |
| `attribution-service` | 8080 | Conversion attribution |
| `audit-service` | 8080 | Compliance audit log |
| `feature-flag-service` | 8080 | Feature flag evaluation |
| `conversion-webhook` | 8080 | Pixel + webhook receiver |
| `admin-backoffice-service` | 8080 | Admin panel backend |

---

## 💻 Local Development

### Prerequisites

```bash
# Required tools
docker      >= 24.0
docker-compose >= 2.20
go          >= 1.24
kubectl     >= 1.28
helm        >= 3.14
terraform   >= 1.6
```

### Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup.git
cd microservice-ecommerce-gcp-cloud-gke-setup

# 2. Start the full local stack (Kafka, PostgreSQL, Redis, all services)
docker compose up -d

# 3. Verify all services are healthy
docker compose ps

# 4. Test the API gateway
curl http://localhost:8080/health
# → {"status":"ok"}

curl http://localhost:8080/ready
# → {"status":"ready"}
```

### Run Tests

```bash
# All tests with race detector and coverage
cd services && go test ./... -race -coverprofile=../coverage.out -covermode=atomic

# View coverage report
go tool cover -html=../coverage.out

# Lint
go install github.com/mgechev/revive@v1.3.7
revive ./...
```

### Build a Single Service

```bash
docker buildx build \
  --build-arg SERVICE_NAME=order-service \
  --build-arg BUILD_SHA=$(git rev-parse HEAD) \
  -t order-service:local \
  -f Dockerfile.services .
```

---

## ☁️ Deployment Guide

### Step 1 — Bootstrap GCP Infrastructure

```bash
# Create the Terraform remote state bucket (one-time)
gcloud storage buckets create gs://ecommerce-tf-state \
  --location=us-central1 \
  --uniform-bucket-level-access \
  --versioning

# Initialise and apply
cd infra/terraform/gke
terraform init
terraform plan -var="project_id=YOUR_PROJECT_ID"
terraform apply -var="project_id=YOUR_PROJECT_ID" -auto-approve
```

### Step 2 — Bootstrap ArgoCD

```bash
# Get credentials for the new cluster
gcloud container clusters get-credentials ecommerce-cell-1 \
  --region us-central1 --project YOUR_PROJECT_ID

# Install ArgoCD
kubectl create namespace platform-argocd
kubectl apply -n platform-argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Bootstrap the App-of-Apps root application
kubectl apply -f argocd/root/application.yaml
```

### Step 3 — CI/CD Secrets

Configure the following secrets in **GitHub → Settings → Secrets and variables → Actions**:

| Secret | Description |
|---|---|
| `GCP_PROJECT_ID` | Your GCP project ID |
| `PAT_TOKEN` | GitHub Personal Access Token (write:packages scope) |
| `SONAR_TOKEN` | SonarQube project token |
| `SONAR_HOST_URL` | SonarQube server URL |

### Step 4 — ArgoCD Application Secrets

```bash
# Create alertmanager secrets (Slack + PagerDuty)
kubectl create secret generic alertmanager-secrets \
  --namespace platform-observability \
  --from-literal=SLACK_WEBHOOK_URL=https://hooks.slack.com/services/... \
  --from-literal=SLACK_SLO_WEBHOOK_URL=https://hooks.slack.com/services/... \
  --from-literal=PAGERDUTY_ROUTING_KEY=your-routing-key
```

---

## 🔐 Environment Variables

### Shared (all services)

| Variable | Default | Description |
|---|---|---|
| `ENV` | `production` | Deployment environment (`dev`/`staging`/`prod`) |
| `LOG_LEVEL` | `info` | Log level (`debug`/`info`/`warn`/`error`) |
| `SERVICE_VERSION` | `unknown` | Semantic version injected at build time |
| `BUILD_SHA` | `unknown` | Git commit SHA injected at build time |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel-collector...:4317` | OTel collector gRPC endpoint |
| `OTEL_INSECURE` | `false` | Set `true` only in local dev to skip TLS |

### Auth Service

| Variable | Source | Description |
|---|---|---|
| `DB_HOST` | ConfigMap | PostgreSQL hostname |
| `DB_PORT` | ConfigMap | PostgreSQL port (5432) |
| `DB_NAME` | ConfigMap | Database name |
| `DB_USER` | Secret `auth-service-secrets` | Database username |
| `DB_PASSWORD` | Secret `auth-service-secrets` | Database password |
| `JWT_SECRET` | Secret `auth-jwt-secret` | JWT signing key |
| `REDIS_HOST` | ConfigMap | Redis hostname |
| `REDIS_PASSWORD` | Secret `auth-redis-secret` | Redis auth password |

### Order Service

| Variable | Source | Description |
|---|---|---|
| `KAFKA_BROKERS` | ConfigMap | Comma-separated Kafka broker list |

---

## 📡 API Reference

All endpoints are prefixed `/v1` through the API Gateway at port 8080.

### Health / Observability

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe — always 200 if process is alive |
| `GET` | `/ready` | Readiness probe — 200 when fully initialised |
| `GET` | `/metrics` | Prometheus metrics endpoint |

### Auth (`/v1/auth`)

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/signup` | Register a new user |
| `POST` | `/v1/login` | Authenticate, returns JWT |
| `GET` | `/v1/validate` | Validate a JWT token |

### Orders (`/v1/orders`)

| Method | Path | Headers | Description |
|---|---|---|---|
| `POST` | `/v1/orders` | `X-Idempotency-Key` | Create an order (idempotent) |
| `GET` | `/v1/orders` | - | List all orders |
| `GET` | `/v1/orders/{id}` | - | Get order by ID |

### Error Response Format

```json
{
  "error": {
    "code": "ORDER_NOT_FOUND",
    "message": "The order 'abc-123' was not found.",
    "request_id": "req-20260320191555.123456789"
  }
}
```

---

## 📈 Scaling Strategy

### Horizontal Pod Autoscaling

All stateless services scale on CPU (70%) and Memory (80%):

| Service | Min | Max | Scale-Up Window | Scale-Down Cooldown |
|---|---|---|---|---|
| `api-gateway` | 2 | 20 | 60s | 5 min |
| `order-service` | 2 | 15 | Immediate | 5 min |
| `auth-service` | 2 | 10 | 30s | 5 min |
| `cart-service` | 2 | 10 | 60s | 5 min |

### Cell-Based Global Scaling

Each **Cell** = 1 GKE Autopilot cluster handling ~500k concurrent users.  
Add cells by expanding `var.cells` in Terraform and ArgoCD auto-syncs all manifests.

```hcl
# Add a new EMEA cell to infra/terraform/gke/variables.tf
cells = {
  "cell-1" = { cell_id = 1, region = "us-central1" }
  "cell-2" = { cell_id = 2, region = "europe-west4" }  # ← new EMEA cell
}
```

---

## 🛡️ SLOs & SLAs

| SLO | Target | Alert Window |
|---|---|---|
| API Availability | 99.9% | 1h (fast-burn) / 6h (slow-burn) |
| P95 Latency | < 500ms | 1h fast-burn |
| P99 Latency | < 1s | 5m threshold |
| Kafka Consumer Lag | < 10k | 5m threshold |

### Alerting Routing

```
CRITICAL (SLO fast-burn, crash loop, endpoint down)  →  PagerDuty + Slack #slo-burn
WARNING  (SLO slow-burn, high CPU/memory, lag)        →  Slack #alerts-warning
```

---

## 🔁 Failure Handling & Runbooks

### Circuit Breaker

All outbound service calls use the built-in circuit breaker (`shared-lib/pkg/resilience`):

- **CLOSED** → normal operation
- **OPEN** → fast-fail after 5 consecutive failures; returns `503` to caller
- **HALF-OPEN** → single probe after 30s timeout; recovers to CLOSED on success

### Retry Policy

```go
resilience.Do(ctx, resilience.RetryConfig{
    MaxAttempts:  3,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     5 * time.Second,
    Multiplier:   2.0,
    Jitter:       0.2,
}, callDownstream)
```

### Graceful Shutdown

All services handle `SIGTERM` with a **30-second drain window**:
1. `/ready` returns `503` immediately → load balancer stops routing new traffic
2. In-flight requests are allowed to complete (max 30s)
3. Kafka producer flushes pending messages
4. OTel tracer flushes pending spans
5. Process exits 0

### Common Runbooks

| Scenario | Runbook Link |
|---|---|
| High error rate | `https://wiki.internal/runbooks/high-error-rate` |
| High latency | `https://wiki.internal/runbooks/high-latency` |
| SLO fast-burn | `https://wiki.internal/runbooks/slo-fast-burn` |
| Pod crash loop | `https://wiki.internal/runbooks/pod-crash-loop` |
| Endpoint down | `https://wiki.internal/runbooks/endpoint-down` |

---

## 🔒 Security Architecture

| Control | Implementation |
|---|---|
| **Zero Trust Networking** | Istio mTLS + NetworkPolicy default-deny |
| **Pod Security** | K8s restricted PSS (no root, no privilege escalation) |
| **Image Supply Chain** | Cosign keyless signing + Kyverno admission policy |
| **Container Scanning** | Trivy (FS scan on PR, container image scan post-push) |
| **Secrets Management** | External Secrets Operator → GCP Secret Manager |
| **Workload Identity** | GKE Workload Identity (no service account keys) |
| **etcd Encryption** | CMEK via Cloud KMS (configurable per cell) |
| **OIDC CI Auth** | GitHub Actions Workload Identity — no long-lived keys |

---

## 🧪 Testing

```bash
# Unit + integration tests with race detector
cd services
go test ./... -race -count=1

# Coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out

# CI enforces ≥80% total coverage
```

### Test Coverage by Package

| Package | What's Tested |
|---|---|
| `shared-lib/pkg/middleware` | Chain order, Metrics recording, Rate limiter (allow/block), RequestID |
| `shared-lib/pkg/resilience` | Retry (5 cases), Circuit breaker (6 state transitions) |

---

## 🤝 Contributing

1. Fork → feature branch → PR to `main`
2. All PRs require: tests pass + coverage ≥80% + Trivy clean + SonarQube gate
3. Staging deploys automatically on merge; prod requires manual approval

---

<div align="center">
Built with ❤️ for production scale · Go 1.24 · GKE Autopilot · ArgoCD GitOps
</div>
