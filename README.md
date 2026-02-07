# 🛒 Cloud-Native E-Commerce Platform

[![Kubernetes](https://img.shields.io/badge/Kubernetes-GKE-326CE5?style=for-the-badge&logo=kubernetes)](https://cloud.google.com/kubernetes-engine)
[![Terraform](https://img.shields.io/badge/Infrastructure-Terraform-7B42BC?style=for-the-badge&logo=terraform)](https://www.terraform.io/)
[![Istio](https://img.shields.io/badge/Service_Mesh-Istio-466BB0?style=for-the-badge&logo=istio)](https://istio.io/)
[![Argo CD](https://img.shields.io/badge/GitOps-Argo_CD-EF7B4D?style=for-the-badge&logo=argo)](https://argo-cd.readthedocs.io/)
[![Go](https://img.shields.io/badge/Backend-Go-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![HTML5](https://img.shields.io/badge/Frontend-HTML5-E34F26?style=for-the-badge&logo=html5)](https://developer.mozilla.org/en-US/docs/Web/HTML)
[![Metabase](https://img.shields.io/badge/Analytics-Metabase-509EE3?style=for-the-badge&logo=metabase)](https://www.metabase.com/)
![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions)

> **Enterprise-ready microservices architecture on Google Cloud Platform.**  
> Hardened for production resiliency, security, and scalability.

---

## 🏗️ Architecture Overview

This platform uses a modern, distributed architecture designed for **high availability**, **observability**, and **secure-by-default operations**.

### Core Stack

- **Compute:** Google Kubernetes Engine (GKE – Autopilot / Standard)
- **Infrastructure:** Terraform (Modular, Validated)
- **Networking:** Istio Service Mesh (mTLS, Gateway API, VirtualServices)
- **Delivery:** Argo CD (GitOps Controller)
- **Data Layer:**
  - PostgreSQL (HA – Bitnami)
  - Redis (Caching – Bitnami)
  - Kafka (Strimzi Operator + KRaft mode)
- **Analytics:**
  - Processing: Apache Spark (K8s) + Airflow
  - Visualization: Metabase (Embedded via signed JWT)
- **Observability:** Prometheus, Grafana, Jaeger, Kiali (OpenTelemetry compatible)

---

## 🚀 Getting Started Guide

### Prerequisites

Make sure the following tools are installed locally:

- [Go 1.24+](https://go.dev/dl/)
- [Docker](https://www.docker.com/)
- [Terraform 1.6+](https://developer.hashicorp.com/terraform/install)
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)

---

## Phase 1: Google Cloud & GitHub Setup

### Google Cloud Setup

1. **Create a GCP Project**
2. **Enable required APIs**
   ```bash
   gcloud services enable \
     compute.googleapis.com \
     container.googleapis.com \
     iam.googleapis.com \
     cloudresourcemanager.googleapis.com \
     artifactregistry.googleapis.com
   ```
