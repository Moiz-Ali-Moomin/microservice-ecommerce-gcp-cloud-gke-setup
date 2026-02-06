FROM apache/airflow:2.9.0

# 1. Switch to root to ensure permissions/system updates (Standard Pattern)
USER root

# Install OpenJDK-17 (Necessary for Spark)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    openjdk-17-jre-headless \
    && apt-get autoremove -yqq --purge \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set JAVA_HOME
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64

# 2. Switch back to airflow user for python handling
USER airflow

# 3. Install Provider with clean environment
# We avoid constraints here to let pip resolve, but explicit user ensures PATH adherence
RUN pip install --no-cache-dir apache-airflow-providers-apache-spark

# 4. Explicitly set PATH to include Airflow's local bin (Critical fix)
ENV PATH="/home/airflow/.local/bin:${PATH}"
