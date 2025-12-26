# Prerequisites

Before installing the Pingora Gateway Controller, ensure you have
the following prerequisites in place.

## Kubernetes Cluster

You need a Kubernetes cluster with:

- Kubernetes version 1.25 or later
- `kubectl` configured to access the cluster
- Helm 3.x installed

## Gateway API CRDs

The controller requires Gateway API Custom Resource Definitions (CRDs) to be
installed in your cluster:

```bash
kubectl apply --filename https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml
```

!!! tip "Version Compatibility"

    The controller is tested with Gateway API v1.4.1. Using older versions
    may result in missing features or compatibility issues.

Verify the CRDs are installed:

```bash
kubectl get crd gateways.gateway.networking.k8s.io
kubectl get crd httproutes.gateway.networking.k8s.io
kubectl get crd grpcroutes.gateway.networking.k8s.io
```

## Resource Requirements

The controller and proxy have minimal resource requirements:

| Component | CPU Request | Memory Request | CPU Limit | Memory Limit |
|-----------|-------------|----------------|-----------|--------------|
| Controller | 100m | 128Mi | 200m | 256Mi |
| Proxy (per replica) | 100m | 128Mi | 500m | 512Mi |

## Network Requirements

The controller requires:

- Access to Kubernetes API server (in-cluster or kubeconfig)
- Network connectivity to Pingora proxy pods (gRPC on port 50051)

The proxy requires:

- Network connectivity to backend services
- Inbound connectivity from clients (HTTP/HTTPS)

## Optional: Prometheus Operator

For metrics collection, the Prometheus Operator enables automatic service discovery:

```bash
kubectl apply --filename https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml
```

See [Monitoring](../guides/monitoring.md) for detailed setup instructions.

## Next Steps

Once you have all prerequisites in place, proceed to [Installation](installation.md).
