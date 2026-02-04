#!/usr/bin/env bash
set -e

docker compose up -d postgres redis
docker compose up -d kafka
docker compose up -d otel-collector
docker compose up -d tempo loki prometheus grafana
docker compose up -d airflow-init
docker compose up -d airflow-webserver airflow-scheduler
docker compose up -d
