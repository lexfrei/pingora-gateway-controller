# Installation

This guide covers installing the Pingora Gateway Controller using Helm.

## Helm Installation

Helm is the recommended installation method. It handles CRD installation,
RBAC setup, and provides a simple upgrade path.

### Basic Installation

```bash
helm install pingora-gateway-controller \
  oci://ghcr.io/lexfrei/pingora-gateway-controller/chart \
  --namespace pingora-system \
  --create-namespace
```

This installs both the controller and the Pingora proxy with default settings.

### Installation with Values File

Create a `values.yaml` file for customization:

```yaml
# Controller settings
controller:
  gatewayClassName: "pingora"
  logLevel: "info"
  logFormat: "json"

# Enable leader election for HA deployments
leaderElection:
  enabled: true

# Multiple controller replicas for high availability
replicaCount: 2

# Proxy settings
proxy:
  enabled: true
  replicaCount: 3
  resources:
    limits:
      cpu: 1
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi

# Enable Prometheus ServiceMonitor
serviceMonitor:
  enabled: true
```

Then install:

```bash
helm install pingora-gateway-controller \
  oci://ghcr.io/lexfrei/pingora-gateway-controller/chart \
  --namespace pingora-system \
  --create-namespace \
  --values values.yaml
```

## Verify Installation

Check that the controller is running:

```bash
kubectl get pods --namespace pingora-system
```

Expected output:

```text
NAME                                                      READY   STATUS    RESTARTS   AGE
pingora-gateway-controller-7d8f9b6c5d-x2j9k               1/1     Running   0          30s
pingora-gateway-controller-proxy-5c4d8b7f6c-m8n3l         1/1     Running   0          30s
pingora-gateway-controller-proxy-5c4d8b7f6c-k9p2q         1/1     Running   0          30s
```

Check GatewayClass:

```bash
kubectl get gatewayclass pingora
```

Expected output:

```text
NAME      CONTROLLER                              ACCEPTED   AGE
pingora   pingora.k8s.lex.la/gateway-controller   True       30s
```

Check Gateway:

```bash
kubectl get gateway --namespace pingora-system
```

## Upgrading

To upgrade to a newer version:

```bash
helm upgrade pingora-gateway-controller \
  oci://ghcr.io/lexfrei/pingora-gateway-controller/chart \
  --namespace pingora-system \
  --values values.yaml
```

## Uninstalling

To remove the controller:

```bash
helm uninstall pingora-gateway-controller \
  --namespace pingora-system
```

!!! warning "Cleanup"

    Uninstalling the Helm release will remove the controller and proxy pods.
    Gateway and HTTPRoute resources in other namespaces will remain but
    become non-functional. Clean them up if no longer needed.

## Manual Installation

For environments where Helm is not available, see
[Manual Installation](../operations/manual-installation.md) for raw
Kubernetes manifests.

## Next Steps

After installation, proceed to [Quick Start](quickstart.md) to create your
first HTTPRoute.
