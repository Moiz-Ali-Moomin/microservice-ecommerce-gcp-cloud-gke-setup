# 🛒 Enterprise Cloud-Native E-Commerce Platform
### Production-Grade Microservices on GKE with Terraform, GitOps, Service Mesh & Data Engineering Pipeline

![Build Status](https://img.shields.io/github/actions/workflow/status/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/?style=for-the-badge&logo=github)
![Go Version](https://img.shields.io/github/go-mod/go-version/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup?style=for-the-badge&logo=go)
![Kubernetes](https://img.shields.io/badge/Orchestration-GKE-326CE5?style=for-the-badge&logo=kubernetes)
![Terraform](https://img.shields.io/badge/IaC-Terraform-7B42BC?style=for-the-badge&logo=terraform)
![Istio](https://img.shields.io/badge/Service_Mesh-Istio-466BB0?style=for-the-badge&logo=istio)
![Argo CD](https://img.shields.io/badge/GitOps-Argo_CD-EF7B4D?style=for-the-badge&logo=argo)
![Kafka](https://img.shields.io/badge/Event_Bus-Kafka-231F20?style=for-the-badge&logo=apachekafka)
![Spark](https://img.shields.io/badge/Big_Data-Spark-E25A1C?style=for-the-badge&logo=apachespark)
![Airflow](https://img.shields.io/badge/Workflow-Airflow-017CEE?style=for-the-badge&logo=apacheairflow)

---

## 📖 Table of Contents

1. [Executive Summary](#-executive-summary)
2. [The "Why": Business Case](#-the-why-business-case--problem-statement)
3. [The "Where": Infrastructure Topology](#-the-where-infrastructure-topology)
4. [The "How": Architecture & Tech Stack](#-the-how-architecture--tech-stack)
5. [The "Who": Roles & Responsibilities](#-the-who-roles--responsibilities)
6. [Enterprise FAQ & Architectural Decisions](#-enterprise-faq--architectural-decisions)
7. [Enterprise Deployment Guide](#-enterprise-gke-platform-deployment-guide)
8. [Verification](#-final-verification-checklist)

---

## 🎯 Executive Summary

This repository represents a **holistic Platform Engineering initiative**. It is not merely a collection of microservices; it is a demonstration of a scalable, secure, and observable ecosystem designed to handle high-concurrency e-commerce workloads.

It bridges the gap between **Application Development** (Go/Gin microservices), **Infrastructure Operations** (Terraform/GKE), and **Data Engineering** (Kafka/Spark/Airflow), unified under a GitOps methodology.

---

## 💡 The "Why": Business Case & Problem Statement

| Challenge | The Solution implemented in this Repo |
| :--- | :--- |
| **Monolithic Bottlenecks** | Decoupled **18+ microservices** allow independent scaling of `cart-service` during flash sales without scaling `user-service`. |
| **Data Silos** | Real-time **Kafka** event streaming combined with **Spark** batch processing ensures marketing and inventory data are consistent. |
| **Operational Toil** | **Terraform** provisions infrastructure idempotently; **Argo CD** ensures the cluster state matches git automatically. |
| **Security Risks** | **Istio** provides Zero-Trust mTLS between services; **External Secrets** prevents credentials from ever touching the git repo. |
| **Blind Spots** | A full **Observability Stack** (Loki/Tempo/Prometheus) provides "Golden Signal" monitoring for instant incident resolution. |

---

## 📍 The "Where": Infrastructure Topology

The platform resides on **Google Cloud Platform (GCP)**, specifically within a Virtual Private Cloud (VPC).

* **Compute Layer:** Google Kubernetes Engine (GKE) - *Regional Cluster*.
* **Data Layer:** Cloud SQL (Postgres), Memorystore (Redis), GCS (Terraform State).
* **Secret Layer:** Google Secret Manager (synced to K8s via External Secrets Operator).
* **Network Layer:**
    * **Public Subnet:** Load Balancers & Istio Ingress Gateway.
    * **Private Subnet:** GKE Nodes, Databases (No public IPs).

---

## ⚙️ The "How": Architecture & Tech Stack

### Traffic Flow Diagram

```mermaid
graph TD
    User((User)) -->|HTTPS| GLB[Google Load Balancer]
    GLB --> Istio[Istio Ingress Gateway]

    subgraph "Kubernetes Cluster (GKE)"
        Istio -->|mTLS| Gateway[API Gateway]

        %% Core Domain
        Gateway --> Auth[Auth Service]
        Gateway --> Product[Product Service]
        Gateway --> Cart[Cart Service]
        Gateway --> Order[Order Service]
        Gateway --> Offer[Offer Service]

        %% Data Ingestion
        Gateway --> Ingest[Analytics Ingest]
        Gateway --> Webhook[Conversion Webhook]

        %% Async Backbone
        Order --> Kafka{Kafka Cluster}
        Ingest --> Kafka

        %% Consumers
        Kafka --> Audit[Audit Service]
        Kafka --> Notif[Notification Service]
        Kafka --> Reporting[Reporting Service]
        
        %% Data Ops
        Airflow[Airflow Orchestrator] -->|Trigger| Spark[Spark Jobs]
        Spark -->|Read| Kafka
        Spark -->|Write| AnalyticsDB[(Postgres Analytics)]
        
        AnalyticsDB --> Metabase[Metabase BI]
    end
👥 The "Who": Roles & Responsibilities
This platform is designed to support cross-functional teams.

Platform Engineer: Manages infra/terraform, GKE, Istio, and GitOps pipelines.

Backend Developer: Owns services/*, utilizing the shared-lib for standardized logging and tracing.

Data Engineer: Manages infra/airflow DAGs, Spark jobs, and Kafka schemas.

Security Engineer: Audits IAM roles, SonarQube gates, and Istio policies.

🏛️ Enterprise FAQ & Architectural Decisions
Q: How are secrets managed securely?
A: We use External Secrets Operator (ESO). Secrets are stored in Google Secret Manager (GSM) and injected into Kubernetes only at runtime. No .env files or base64 secrets exist in the git repository.

Q: How does the system handle Flash Sales (High Load)?
A:

HPA (Horizontal Pod Autoscaler): Scales pods based on CPU/Memory/Custom Metrics.

Cluster Autoscaler: Provisions new GKE nodes when pod capacity is reached.

Redis Caching: product-service and cart-service heavily utilize Redis to reduce DB load.

Q: What is the Disaster Recovery (DR) strategy?
A: Infrastructure is defined in Terraform (IaC). In a catastrophic failure, we can spin up a new cluster in a different region using the same scripts. Database backups are handled by Cloud SQL automated backups.

Q: Is the communication secure?
A: Yes. Istio Service Mesh enforces strict mTLS (mutual TLS) between all microservices. Services cannot talk to each other unless explicitly allowed by AuthorizationPolicies.

🚀 Enterprise GKE Platform Deployment Guide
Follow these 14 steps sequentially to provision the entire platform from scratch.

Step 1 — Init GCP
Initialize your Google Cloud environment and enable required APIs.

Bash
gcloud auth login

gcloud config set project ecommerce-microservice-53

gcloud services enable \
    container.googleapis.com \
    artifactregistry.googleapis.com \
    cloudbuild.googleapis.com \
    secretmanager.googleapis.com \
    iam.googleapis.com
Step 2 — Terraform Infrastructure
Provision the base infrastructure components: State Backend, Identity, and Secrets.

Bash
# 2.1 Backend
cd infra/terraform/gcp/bootstrap-backend
terraform init
terraform apply -auto-approve

# 2.2 Identity
cd ../bootstrap-identity
terraform init
terraform apply -auto-approve

# 2.3 Secrets Manager
cd ../secrets
terraform init
terraform apply -auto-approve
Step 3 — Provision GKE Cluster
Deploy the Kubernetes cluster and node pools.

Bash
cd ../../../gke

terraform init
terraform apply -auto-approve

# Connect to the new cluster
gcloud container clusters get-credentials ecommerce-platform-prod --region us-central1

# Verify nodes
kubectl get nodes
Step 4 — Namespaces
Create the logical isolation boundaries for the platform.

Bash
kubectl apply -f infra/namespaces/

kubectl get ns
Step 5 — External Secrets
Install the operator to sync Google Secret Manager secrets into Kubernetes.

Bash
helm upgrade --install external-secrets infra/external-secrets \
    -n external-secrets \
    --create-namespace \
    --wait

# Apply ClusterSecretStore and ExternalSecrets
kubectl apply -f infra/secrets/external-secrets.yaml
kubectl apply -f infra/app-secrets/

# Verify secrets are synced
kubectl get externalsecret -A
Step 6 — PostgreSQL and Redis
Deploy the stateful data layer.

Bash
helm repo add bitnami [https://charts.bitnami.com/bitnami](https://charts.bitnami.com/bitnami)
helm repo update

# PostgreSQL
helm upgrade --install postgres bitnami/postgresql \
    -n data-postgres \
    -f infra/databases/postgres/values.yaml \
    --wait

# Redis
helm upgrade --install redis bitnami/redis \
    -n data-redis \
    -f infra/databases/redis/values.yaml \
    --wait
Step 7 — Kafka (Event Backbone)
Deploy the Strimzi operator and the Kafka cluster.

Bash
# Install Operator
helm upgrade --install strimzi strimzi/strimzi-kafka-operator \
    -n kafka \
    --create-namespace \
    --wait

# Deploy Cluster & Topics
kubectl apply -f infra/kafka/storage/
kubectl apply -f infra/kafka/cluster/
kubectl apply -f infra/kafka/topics/
Step 8 — Observability Stack
Deploy Prometheus, Grafana, and related monitoring tools.

Bash
helm upgrade --install kube-prometheus-stack \
    prometheus-community/kube-prometheus-stack \
    -n platform-observability \
    --create-namespace \
    --wait
Step 9 — Istio Service Mesh
Install the service mesh control plane and configure ingress gateways.

Bash
# Install Istio
istioctl install --set profile=default --set components.cni.enabled=false -y

# Enable Sidecar Injection
kubectl label namespace apps-public istio-injection=enabled --overwrite
kubectl label namespace apps-core istio-injection=enabled --overwrite
kubectl label namespace apps-async istio-injection=enabled --overwrite

# Deploy Gateways and Security Policies
kubectl apply -f infra/istio/gateway-workload/
kubectl apply -f infra/istio/ingress/
kubectl apply -f infra/istio/security/
Step 10 — Build and Push Images
Build all microservices using Google Cloud Build and push to Artifact Registry.

Bash
# Triggers the build for all services defined in cloudbuild.yaml
gcloud builds submit --config cloudbuild.yaml .

# Verify images
gcloud artifacts docker images list \
    us-central1-docker.pkg.dev/ecommerce-microservice-53/ecommerce-repo
Step 11 — Deploy Data Tools (Airflow & Metabase)
Deploy the data engineering orchestration and BI tools.

Bash
# Airflow
helm upgrade --install airflow apache-airflow/airflow \
    -n tooling-airflow \
    -f infra/airflow/values.yaml \
    --wait

# Metabase
helm upgrade --install metabase oci://registry-1.docker.io/bitnamicharts/metabase \
    -n tooling-metabase \
    -f infra/metabase/values.yaml \
    --create-namespace \
    --wait

# Access Metabase UI
kubectl port-forward svc/metabase -n tooling-metabase 3001:3000
Step 12 — Deploy ALL Microservices
Deploy the application layer, categorized by domain.

Core Services
Bash
helm upgrade --install auth-service services/auth-service/helm -n apps-core --wait
helm upgrade --install user-service services/user-service/helm -n apps-core --wait
helm upgrade --install product-service services/product-service/helm -n apps-core --wait
helm upgrade --install order-service services/order-service/helm -n apps-core --wait
helm upgrade --install cart-service services/cart-service/helm -n apps-core --wait
helm upgrade --install feature-flag-service services/feature-flag-service/helm -n apps-core --wait
helm upgrade --install notification-service services/notification-service/helm -n apps-core --wait
helm upgrade --install offer-service services/offer-service/helm -n apps-core --wait
Public Services
Bash
helm upgrade --install api-gateway services/api-gateway/helm -n apps-public --wait
helm upgrade --install storefront-service services/storefront-service/helm -n apps-public --wait
helm upgrade --install landing-service services/landing-service/helm -n apps-public --wait
helm upgrade --install redirect-service services/redirect-service/helm -n apps-public --wait
helm upgrade --install admin-backoffice-service services/admin-backoffice-service/helm -n apps-public --wait
Async Services
Bash
helm upgrade --install analytics-ingest-service services/analytics-ingest-service/helm -n apps-async --wait
helm upgrade --install attribution-service services/attribution-service/helm -n apps-async --wait
helm upgrade --install audit-service services/audit-service/helm -n apps-async --wait
helm upgrade --install reporting-service services/reporting-service/helm -n apps-async --wait
helm upgrade --install conversion-webhook services/conversion-webhook/helm -n apps-async --wait
Step 13 — Deploy Spark Jobs
Submit the data processing jobs to the cluster.

Bash
kubectl apply -f infra/spark/spark-submit.yaml

kubectl get pods -n spark
Step 14 — Final Verification
Ensure the platform is healthy and operational.

Bash
# 1. Check all pods are running
kubectl get pods -A

# 2. Verify Secret Synchronization
kubectl get externalsecret -A

# 3. Check Ingress IP
kubectl get svc istio-ingressgateway -n istio-system

# 4. Check Data Tooling
kubectl get pods -n tooling-airflow
kubectl get pods -n tooling-metabase

# 5. Check Persistence
kubectl get pods -n data-postgres
kubectl get pods -n data-redis
✅ Final Verification Checklist
The platform is considered READY when:

[ ] All Pods are in Running state (No CrashLoopBackOff).

[ ] Istio Ingress Gateway has an assigned External IP.

[ ] All images are successfully present in Artifact Registry.

[ ] Airflow, Spark, and Metabase are accessible.

[ ] External Secrets have status Synced.

📄 License
MIT License
