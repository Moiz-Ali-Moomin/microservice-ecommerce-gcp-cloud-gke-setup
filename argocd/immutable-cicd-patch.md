# CI/CD and GitOps Immutability Patch Recommendations

# 1. GitHub Actions Patch (.github/workflows/ci.yml)
# Prevent image mutability by dropping the `:latest` tag and exclusively using Git SHAs.

```yaml
      - name: Build and Push Docker image
        uses: docker/build-push-action@v4
        with:
          context: ./services/${{ matrix.service }}
          push: true
          # OLD: tags: my-registry/${{ matrix.service }}:latest
          # NEW:
          tags: |
            my-registry/${{ matrix.service }}:${{ github.sha }}
            my-registry/${{ matrix.service }}:v${{ github.run_number }}
```

# 2. ArgoCD Application Manifest Patch
# Ensure tracking matches the immutable tag, configured dynamically via Kustomize or Helm values.

```yaml
source:
  repoURL: https://github.com/my-org/repo
  path: services/api-gateway/helm
  helm:
    parameters:
    # ArgoCD should inject the SHA natively via an image updater
    - name: image.tag
      value: "9a8b7c6d5e4f3a2b1c" 
```

# 3. Argo Rollouts (Canary Deployment Strategy)
# Replace Kubernetes `Deployment` with standard `Rollout` for safe percentage-based progression.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: order-service
spec:
  replicas: 5
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 5m}  # Bake for 5 minutes
      - setWeight: 50
      - pause: {duration: 10m} # Observe error rates
      # Full promotion
```
