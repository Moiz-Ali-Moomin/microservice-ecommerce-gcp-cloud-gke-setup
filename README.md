# 🛒 Cloud-Native E-Commerce Platform

[![Kubernetes](https://img.shields.io/badge/Kubernetes-GKE-326CE5?style=for-the-badge&logo=kubernetes)](https://cloud.google.com/kubernetes-engine)
[![Terraform](https://img.shields.io/badge/Infrastructure-Terraform-7B42BC?style=for-the-badge&logo=terraform)](https://www.terraform.io/)
[![Istio](https://img.shields.io/badge/Service_Mesh-Istio-466BB0?style=for-the-badge&logo=istio)](https://istio.io/)
[![Argo CD](https://img.shields.io/badge/GitOps-Argo_CD-EF7B4D?style=for-the-badge&logo=argo)](https://argo-cd.readthedocs.io/)
[![Go](https://img.shields.io/badge/Backend-Go-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![HTML5](https://img.shields.io/badge/Frontend-HTML5-E34F26?style=for-the-badge&logo=html5)](https://developer.mozilla.org/en-US/docs/Web/HTML)
[![Metabase](https://img.shields.io/badge/Analytics-Metabase-509EE3?style=for-the-badge&logo=metabase)](https://www.metabase.com/)
![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions)

> **Enterprise-Ready Microservices Architecture on Google Cloud Platform.**
> Currently hardened for production resiliency, security, and scalability.

---

## 🏗️ Architecture Overview

This platform uses a modern, distributed architecture designed for high availability and observability.

### **Core Stack**
*   **Compute:** Google Kubernetes Engine (GKE) Autopilot/Standard
*   **Infrastructure:** Terraform (Modular, Validated)
*   **Networking:** Istio Service Mesh (mTLS, Gateway API, VirtualServices)
*   **Delivery:** Argo CD (GitOps Controller)
*   **Data Layer:**
    *   **PostgreSQL:** HA Cluster (Bitnami)
    *   **Redis:** Caching Layer (Bitnami)
    *   **Kafka:** Event Streaming (Strimzi Operator + KRaft Mode)
*   **Analytics:**
    *   **Processing:** Apache Spark (on K8s) & Airflow (Orchestration)
    *   **Visualization:** Metabase (Self-hosted, Embedded via Signed JWT)
*   **Observability:** Prometheus, Grafana, Jaeger, Kiali (OpenTelemetry compatible)

---

## 🚀 Getting Started Guide

Follow these steps to go from zero to a fully running production cluster.

### **Prerequisites**
Ensure you have the following tools installed locally:
*   [Go 1.24+](https://go.dev/dl/)
*   [Docker](https://www.docker.com/) (Desktop or Engine)
*   [Terraform 1.6+](https://developer.hashicorp.com/terraform/install)
*   [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (`gcloud`)
*   [Kubectl](https://kubernetes.io/docs/tasks/tools/)
*   [Helm](https://helm.sh/docs/intro/install/)

---

### **Phase 1: Google Cloud & GitHub Setup**

Before running any code, you need to prepare your environment.

#### **1. Google Cloud Setup**
1.  **Create a Project:** Create a new GCP Project.
2.  **Enable APIs:**
    ```bash
    gcloud services enable compute.googleapis.com \
        container.googleapis.com \
        iam.googleapis.com \
        cloudresourcemanager.googleapis.com \
        artifactregistry.googleapis.com
    ```
3.  **Local Auth:** Authenticate your local terminal to GCP.
    ```bash
    gcloud auth application-default login
    ```

#### **2. GitHub Actions Secrets**
To enable the CI/CD pipelines, go to your GitHub Repo -> **Settings** -> **Secrets and variables** -> **Actions** and add:

| Secret Name | Value | Description |
| :--- | :--- | :--- |
| `GCP_PROJECT_ID` | `your-gcp-project-id` | Target Google Cloud Project ID |
| `SONAR_TOKEN` | `(Generated in Sonar)` | Token for SonarQube analysis |
| `SONAR_HOST_URL` | `https://sonar.example.com` | URL of your SonarQube instance |
| `SONAR_ADMIN_PASSWORD` | `your-secure-password` | Initial Admin password for Sonar Setup |

---

### **Phase 2: Infrastructure Provisioning (Terraform)**

We use a layered Terraform approach for safety.

#### **Step 1: Bootstrap Backend (Remote State)**
This layer creates the GCS bucket for Terraform state locking.
```bash
cd infra/terraform/gcp/bootstrap-backend
terraform init
terraform apply -var="project_id=your-gcp-project-id"
```

#### **Step 2: Bootstrap Identity (IAM & Workload Identity)**
This layer sets up the trust between GitHub Actions and GCP.
```bash
cd infra/terraform/gcp/bootstrap-identity
# Update variables.tf or pass vars via -var
terraform init
terraform apply -var="project_id=your-gcp-project-id"
```

#### **Step 3: Platform Infrastructure (GKE & VPC)**
This creates the actual Kubernetes Cluster and Networking.
```bash
cd ../platform
terraform init
terraform apply -var="project_id=your-gcp-project-id"
```
> **Output:** Note the `cluster_name` and `cluster_endpoint` from the output.

---

### **Phase 3: Kubernetes Deployment (GitOps Bootstrap)**

Once Terraform finishes, configure your local `kubectl`:
```bash
gcloud container clusters get-credentials ecommerce-prod --region us-central1 --project your-gcp-project-id
```

We use **Argo CD** to manage all cluster resources. You only need to install Argo CD once; it will handle the rest.

#### **1. Install Argo CD**
```bash
kubectl create ns platform-argocd
kubectl apply -k argocd/bootstrap/
```

#### **2. Enable the "Root App"**
This single manifest triggers the synchronization of the entire platform (Istio, Data, Apps).
```bash
kubectl apply -f argocd/root/application.yaml
```

#### **3. Monitor Synchronization**
You can watch the pods come up, or port-forward the UI:
```bash
kubectl port-forward svc/argocd-server -n platform-argocd 8080:443
# Login with user: admin, password: (kubectl -n platform-argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
```

---

## 🔄 GitOps Architecture

We use the **App of Apps** pattern to manage the entire cluster state from Git.

### **Responsibility Model**
*   **Terraform:** Manages Cloud Infrastructure (GKE, VPC, IAM).
*   **GitHub Actions:** Builds images, runs tests, and updates Helm Versions in Git.
*   **Argo CD:** Reconciles the state of the cluster with the Git repository. **Zero Touch Prod.**

### **Argo CD Structure**
*   `argocd/projects/`: Defines 4 isolated projects:
    *   **Platform:** High privilege (Namespaces, CRDs, Istio).
    *   **Data:** Stateful workloads (Kafka, Postgres).
    *   **Tooling:** Ops tools (Airflow, SonarQube).
    *   **Applications:** Restricted Microservices (Deployment/Service only).
*   `argocd/apps/`: The actual Application manifests mapping to `infra/` and `services/`.
*   `argocd/root/`: The master controller.

---

## 📂 Project Structure

```bash
ecommerce-platform/
├── resources/          # Application configurations (Kafka Topics, etc.)
│   └── kafka-topics/   # Declarative Kafka Topics
├── argocd/             # GitOps Configuration
│   ├── bootstrap/      # Install manifests
│   ├── projects/       # Security boundaries
│   ├── apps/           # Workload definitions
│   └── root/           # App-of-Apps entrypoint
├── infra/              # Infrastructure as Code
│   ├── airflow/        # Airflow Helm values & DAGs
│   ├── databases/      # Postgres & Redis HA configs
│   ├── istio/          # Service Mesh config (Gateways, Auth)
│   ├── kafka/          # Strimzi Operator & Cluster
│   ├── namespaces/     # Production Namespace Strategy
│   ├── spark/          # Spark Jobs & Operator
│   └── terraform/      # GCP Infrastructure (GKE, VPC)
├── services/           # Microservices (Go)
│   ├── cart-service/
│   ├── product-service/
│   └── payment-service/
└── .github/workflows/  # CI/CD Pipelines
```

---

## 🛡️ Security & Hardening

This platform implements **Defense in Depth**:
1.  **Network Policies:** Strict namespace isolation (e.g., `apps-public` cannot talk to `data-postgres`).
2.  **Workload Identity:** No static keys. Pods authenticate to GCP via Federation.
3.  **mTLS:** Enforced by Istio for all service-to-service communication.
4.  **Least Privilege:** Service Accounts are scoped strictly to their needs.
5.  **GitOps Isolation:** Argo CD Projects prevent apps from touching infra.

---

## 🧪 Development Workflow

### **Local Testing (Docker Compose)**
You can run the logic locally without Kubernetes:
```bash
docker-compose up --build
```

### **CI/CD Pipeline**
1.  **Dev Pushes Code:** GitHub Actions runs Unit Tests & Security Scans.
2.  **Artifact Build:** CI builds and pushes Docker Image to Artifact Registry.
3.  **Git Update:** CI updates the Helm Chart `appVersion` in `services/cart-service/helm/Chart.yaml` (coming soon).

### **GitOps Operations**

*   **Rollbacks:** To roll back a change, `git revert` the commit in this repository. Argo CD will automatically sync the previous state.
*   **Drift Detection:** Argo CD automatically detects if someone manually changes the cluster (e.g., via `kubectl edit`). It will mark the app as "OutOfSync" or auto-correct it if `selfHeal` is enabled.
*   **UI Security:** The Argo CD UI is **internal only**. Do not expose it via LoadBalancer. Use `kubectl port-forward` for administrative access only.
4.  **Argo CD Sync:** Argo CD detects the change in Git and updates the cluster.

---

> Built with ❤️ by the Platform Engineering Team.
