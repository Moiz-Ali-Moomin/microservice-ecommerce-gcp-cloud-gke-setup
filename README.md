# Cloud-Native Enterprise E-Commerce Platform

A production-grade, self-hosted, event-driven e-commerce platform designed for high-scale affiliate marketing and sales.

## 🚀 Overview

This platform is a comprehensive microservices-based system built with Go, deployed on Google Kubernetes Engine (GKE), and leveraging a self-hosted event backbone (Kafka). It is engineered to demonstrate real-world enterprise constraints, such as data sovereignty, custom analytics pipelines, and strict security compliance without relying on managed cloud data services.

## 🏗 System Architecture

See [docs/architecture.md](docs/architecture.md) for a deep dive into:
- Service boundaries
- Event flows
- Data pipeline
- Failure domains

### Core Components
- **Microservices**: Built in Go (1.22+), communicating via gRPC and Kafka.
- **Service Mesh**: Istio for traffic management, mTLS, and observability.
- **Event Bus**: Apache Kafka (Self-hosted on GKE) for asynchronous event processing.
- **Data Pipeline**: Custom ETL using Apache Airflow and Apache Spark reading from GCS buckets.
- **Infrastructure**: GKE Standard, Cloud Storage, PostgreSQL, Redis.
- **Observability**: Prometheus, Grafana, Loki, Tempo.

## 📂 Repository Structure

```
ecommerce-platform/
├── services/           # Go microservices (Domain logic)
├── infra/              # Kubernetes manifests, Helm charts, Terraform
├── docs/               # Architecture, Security, and Operational docs
├── ci/                 # GitLab CI pipeline configuration
└── scripts/            # Utility scripts for dev, test, and ops
```

## 🛠 Tech Stack

- **Languages**: Go, Python (Airflow/Spark)
- **Containerization**: Docker
- **Orchestration**: Kubernetes (GKE)
- **Mesh**: Istio
- **Messaging**: Kafka (Strimzi/Bitnami)
- **CI/CD**: GitLab CI
- **Monitoring**: GLGT Stack (Grafana, Loki, Tempo, Prometheus)

## 🚦 Getting Started

See [docs/architecture.md](docs/architecture.md) for a deep dive into the system design.

To run locally (requires local K8s cluster like Kind or Minikube):
```bash
./scripts/local-dev.sh
```

## 📜 License

Private Enterprise License.
