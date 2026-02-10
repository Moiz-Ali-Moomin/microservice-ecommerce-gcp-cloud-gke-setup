# Kafka Strimzi Deployment - Troubleshooting

## Issue Summary
Strimzi Kafka operator stuck in `CrashLoopBackOff` due to `ClusterRoleBinding` conflicts.

## Root Cause
Conflicting `roleRef` between official and custom Strimzi installs.

## Resolution
1. Delete conflicting `ClusterRoleBinding` named `strimzi-cluster-operator`.
2. Create `kafka` namespace FIRST.
3. Install ONLY from official source: `https://strimzi.io/install/latest?namespace=kafka`.
