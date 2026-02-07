FROM apache/airflow:2.9.0

USER root

# Java required for Spark
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    openjdk-17-jre-headless && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64

USER airflow

# 🔒 Pin provider version compatible with Airflow 2.9.x
RUN pip install --no-cache-dir \
    apache-airflow-providers-apache-spark==4.8.0
