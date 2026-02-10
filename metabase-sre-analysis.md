# SRE Post-Mortem & Remediation: Metabase Deployment

## Root Cause Analysis

### Scheduling Failures
*   **Mismatched Taints/Tolerations**: Pods targeting `workload=analytics:NoSchedule` nodes lacked `tolerations`.
*   **Autoscaler Stagnation**: Cluster Autoscaler refused scale-up for untolerated taints.

### Application Failures
*   **Probe Killing Migrations**: Liveness probes killed pods during Clojure startup/migrations.
*   **Resource Starvation**: Default 1GB limits were insufficient for Metabase JVM.

---

## Production-Grade Manifest (Key Controls)

```yaml
# SRE Migration Safety
startupProbe:
  httpGet:
    path: /api/health
    port: 3000
  failureThreshold: 60
  periodSeconds: 10

# Scheduling Correctness
tolerations:
- key: "workload"
  operator: "Equal"
  value: "analytics"
  effect: "NoSchedule"
```
