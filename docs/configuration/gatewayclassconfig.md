# PingoraConfig CRD

PingoraConfig is a cluster-scoped Custom Resource Definition (CRD) that
configures the connection between the controller and Pingora proxy.

## Overview

The PingoraConfig resource is referenced by GatewayClass via `parametersRef`:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: pingora
spec:
  controllerName: pingora.k8s.lex.la/gateway-controller
  parametersRef:
    group: pingora.k8s.lex.la
    kind: PingoraConfig
    name: pingora-config
```

## Resource Definition

```yaml
apiVersion: pingora.k8s.lex.la/v1alpha1
kind: PingoraConfig
metadata:
  name: pingora-config
spec:
  # Required: gRPC endpoint address
  address: "pingora-proxy.pingora-system.svc.cluster.local:50051"

  # Optional: TLS configuration
  tls:
    enabled: false
    secretRef:
      name: pingora-tls
      namespace: pingora-system
    insecureSkipVerify: false
    serverName: ""

  # Optional: Connection parameters
  connection:
    connectTimeoutSeconds: 5
    requestTimeoutSeconds: 30
    keepaliveTimeSeconds: 30
    maxRetries: 3
    retryBackoffMs: 1000
```

## Specification

### `spec.address`

**Required.** The gRPC endpoint address of the Pingora proxy.

| Field | Type | Description |
|-------|------|-------------|
| `address` | string | Format: `host:port` |

Example:

```yaml
spec:
  address: "pingora-proxy.pingora-system.svc.cluster.local:50051"
```

### `spec.tls`

Optional TLS configuration for the gRPC connection.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | `false` | Enable TLS for gRPC connection |
| `secretRef.name` | string | - | Secret containing TLS certificates |
| `secretRef.namespace` | string | - | Namespace of the Secret |
| `insecureSkipVerify` | boolean | `false` | Skip certificate verification |
| `serverName` | string | - | Override server name for TLS |

!!! danger "insecureSkipVerify"

    Setting `insecureSkipVerify: true` disables certificate verification.
    Only use this for testing environments.

#### TLS Secret Format

The referenced Secret must contain:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pingora-tls
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-certificate>
  tls.key: <base64-encoded-key>
  ca.crt: <base64-encoded-ca-certificate>  # Optional
```

### `spec.connection`

Optional connection parameters for gRPC.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `connectTimeoutSeconds` | int32 | `5` | Connection establishment timeout |
| `requestTimeoutSeconds` | int32 | `30` | Individual request timeout |
| `keepaliveTimeSeconds` | int32 | `30` | Keepalive ping interval |
| `maxRetries` | int32 | `3` | Maximum retry attempts |
| `retryBackoffMs` | int32 | `1000` | Backoff between retries (ms) |

## Status

The controller updates the PingoraConfig status:

```yaml
status:
  conditions:
    - type: Ready
      status: "True"
      reason: Connected
      message: "Successfully connected to Pingora proxy"
  connected: true
  lastSyncTime: "2024-01-15T10:30:00Z"
  configVersion: 42
```

| Field | Description |
|-------|-------------|
| `connected` | Connection to proxy established |
| `lastSyncTime` | Last successful route sync |
| `configVersion` | Current configuration version |

## Examples

### Basic Configuration

Minimal configuration for in-cluster proxy:

```yaml
apiVersion: pingora.k8s.lex.la/v1alpha1
kind: PingoraConfig
metadata:
  name: pingora-config
spec:
  address: "pingora-proxy.pingora-system.svc.cluster.local:50051"
```

### With TLS

Secure connection with mTLS:

```yaml
apiVersion: pingora.k8s.lex.la/v1alpha1
kind: PingoraConfig
metadata:
  name: pingora-config
spec:
  address: "pingora-proxy.pingora-system.svc.cluster.local:50051"
  tls:
    enabled: true
    secretRef:
      name: controller-mtls
      namespace: pingora-system
    serverName: "pingora-proxy"
```

### Custom Timeouts

For high-latency or unreliable networks:

```yaml
apiVersion: pingora.k8s.lex.la/v1alpha1
kind: PingoraConfig
metadata:
  name: pingora-config
spec:
  address: "pingora-proxy.remote-cluster.svc.cluster.local:50051"
  connection:
    connectTimeoutSeconds: 10
    requestTimeoutSeconds: 60
    keepaliveTimeSeconds: 15
    maxRetries: 5
    retryBackoffMs: 2000
```

## Troubleshooting

### Connection Issues

Check PingoraConfig status:

```bash
kubectl get pingoraconfig pingora-config --output yaml
```

Verify proxy is reachable:

```bash
kubectl exec -it deployment/pingora-gateway-controller -- \
  nc -zv pingora-proxy.pingora-system.svc.cluster.local 50051
```

### TLS Errors

Verify Secret exists and has correct keys:

```bash
kubectl get secret pingora-tls --namespace pingora-system --output yaml
```

Check certificate validity:

```bash
kubectl get secret pingora-tls --namespace pingora-system \
  --output jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text -noout
```

## Next Steps

- Learn about [Gateway API](../gateway-api/index.md) resources
- Set up [Monitoring](../guides/monitoring.md) for production
