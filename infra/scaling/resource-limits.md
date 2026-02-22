# Resource Limit Recommendations per Service Tier

A PRR-compliant cluster must strictly define Requests and Limits to prevent Noisy Neighbor scenarios and enable deterministic HPA scale-outs.

## Tier 1: High Throughput Edge / Gateway Services
**Services:** `api-gateway`, `storefront-web`, `auth-service`
* **CPU Request:** `200m`
* **CPU Limit:** `1000m`
* **Memory Request:** `128Mi`
* **Memory Limit:** `512Mi`
* **HPA Target:** `min: 3`, `max: 50`

## Tier 2: Transactional Core (DB Heavy)
**Services:** `order-service`, `cart-service`, `user-service`, `offer-service`, `product-service`
* **CPU Request:** `100m`
* **CPU Limit:** `500m`
* **Memory Request:** `64Mi`  (Go is highly memory efficient)
* **Memory Limit:** `256Mi`
* **HPA Target:** `min: 2`, `max: 30`

## Tier 3: Async Event Processors (Kafka Heavy)
**Services:** `notification-service`, `analytics-ingest-service`, `audit-service`
* **CPU Request:** `50m`
* **CPU Limit:** `250m`
* **Memory Request:** `64Mi`
* **Memory Limit:** `256Mi`
* **Kafka Scaling:** Partition count must match Max HPA replicas (e.g., 30 partitions for `page.viewed` if maxReplicas is 30) to ensure consumer parallelism.

## Tier 4: JVM / Big Data (Spark ELT)
**Services:** `funnel-analysis` Spark Driver and Executors
* **Driver Limits:** `cores: 1`, `memory: 1G`
* **Executor Limits:** `cores: 1`, `memory: 2G`
* **Instances:** Dynamically allocated based on Kafka offset lag (Spark Dynamic Allocation).
