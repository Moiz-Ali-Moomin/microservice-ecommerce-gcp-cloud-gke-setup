# Istio Base ConfigurationThis directory contains the base configuration for Istio.> **Note**: This setup assumes the Istio control plane (`istiod`) is already installed and managed separately (e.g., via a platform team or a separate ArgoCD application).The configurations in `infra/istio/` are strictly for the **data plane** (ingress gateway) and **traffic management/security** policies.## Responsibilities- **Bootstrap**: No CRDs or operator configurations are present here to avoid conflicts.- **GitOps**: This structure supports a clean separation of concerns.


istioctl install --set profile=default -y

