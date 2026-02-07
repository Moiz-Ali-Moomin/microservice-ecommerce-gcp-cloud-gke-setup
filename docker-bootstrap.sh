#!/usr/bin/env bash
set -e

# Start core data services first
docker compose up -d postgres redis

# Start streaming platform
docker compose up -d kafka

# Start observability stack
docker compose up -d otel-collector
docker compose up -d tempo loki prometheus grafana

# Start Airflow (init first, then services)
docker compose up -d airflow-init
docker compose up -d airflow-webserver airflow-scheduler

# Start everything else
docker compose up -d