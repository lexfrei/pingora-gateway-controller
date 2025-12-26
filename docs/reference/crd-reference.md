# CRD Reference

Complete specification of the PingoraConfig Custom Resource Definition.

## PingoraConfig

PingoraConfig is a cluster-scoped resource that configures the connection
between the controller and Pingora proxy.

### API Version

```yaml
apiVersion: pingora.k8s.lex.la/v1alpha1
kind: PingoraConfig
```

### Metadata

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Resource name (cluster-scoped) |
| `labels` | map | Standard Kubernetes labels |
| `annotations` | map | Standard Kubernetes annotations |

### Spec

#### spec.address

**Required.** The gRPC endpoint address of the Pingora proxy.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `address` | string | Yes | Format: `host:port` |

Example:

```yaml
spec:
  address: "pingora-proxy.pingora-system.svc.cluster.local:50051"
```

#### spec.tls

Optional TLS configuration for the gRPC connection.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | `false` | Enable TLS |
| `secretRef` | object | - | Reference to TLS Secret |
| `insecureSkipVerify` | boolean | `false` | Skip certificate verification |
| `serverName` | string | - | Override server name |

##### spec.tls.secretRef

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Secret name |
| `namespace` | string | No | Secret namespace |

The referenced Secret must contain:

- `tls.crt` - TLS certificate
- `tls.key` - TLS private key
- `ca.crt` - CA certificate (optional)

Example:

```yaml
spec:
  tls:
    enabled: true
    secretRef:
      name: pingora-mtls
      namespace: pingora-system
    serverName: "pingora-proxy"
```

#### spec.connection

Optional connection parameters for gRPC.

| Field | Type | Default | Min | Description |
|-------|------|---------|-----|-------------|
| `connectTimeoutSeconds` | int32 | `5` | 1 | Connection timeout |
| `requestTimeoutSeconds` | int32 | `30` | 1 | Request timeout |
| `keepaliveTimeSeconds` | int32 | `30` | 10 | Keepalive interval |
| `maxRetries` | int32 | `3` | 0 | Max retry attempts |
| `retryBackoffMs` | int32 | `1000` | 100 | Retry backoff (ms) |

Example:

```yaml
spec:
  connection:
    connectTimeoutSeconds: 10
    requestTimeoutSeconds: 60
    keepaliveTimeSeconds: 15
    maxRetries: 5
    retryBackoffMs: 2000
```

### Status

The controller updates the status subresource.

| Field | Type | Description |
|-------|------|-------------|
| `conditions` | []Condition | Standard Kubernetes conditions |
| `connected` | boolean | Connection established |
| `lastSyncTime` | Time | Last successful sync |
| `configVersion` | uint64 | Current config version |

#### Conditions

| Type | Reason | Description |
|------|--------|-------------|
| `Ready` | `Connected` | Successfully connected to proxy |
| `Ready` | `ConnectionFailed` | Failed to connect |
| `Ready` | `ConfigurationInvalid` | Invalid configuration |

### Short Name

The CRD registers the short name `pgconfig`:

```bash
kubectl get pgconfig
```

### Print Columns

| Name | Path | Description |
|------|------|-------------|
| Address | `.spec.address` | Proxy address |
| TLS | `.spec.tls.enabled` | TLS enabled |
| Connected | `.status.connected` | Connection status |
| Age | `.metadata.creationTimestamp` | Resource age |

## Complete Example

```yaml
apiVersion: pingora.k8s.lex.la/v1alpha1
kind: PingoraConfig
metadata:
  name: pingora-config
  labels:
    app.kubernetes.io/name: pingora-gateway-controller
    app.kubernetes.io/managed-by: Helm
spec:
  # Required: Proxy gRPC address
  address: "pingora-proxy.pingora-system.svc.cluster.local:50051"

  # Optional: TLS configuration
  tls:
    enabled: true
    secretRef:
      name: pingora-mtls
      namespace: pingora-system
    insecureSkipVerify: false
    serverName: "pingora-proxy"

  # Optional: Connection parameters
  connection:
    connectTimeoutSeconds: 5
    requestTimeoutSeconds: 30
    keepaliveTimeSeconds: 30
    maxRetries: 3
    retryBackoffMs: 1000
```

## GatewayClass Integration

PingoraConfig is referenced by GatewayClass via `parametersRef`:

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

## Validation

The CRD includes validation rules:

| Field | Validation |
|-------|------------|
| `spec.address` | Required, min length 1 |
| `spec.tls.secretRef.name` | Required if secretRef specified, min length 1 |
| `spec.connection.connectTimeoutSeconds` | Minimum 1 |
| `spec.connection.requestTimeoutSeconds` | Minimum 1 |
| `spec.connection.keepaliveTimeSeconds` | Minimum 10 |
| `spec.connection.maxRetries` | Minimum 0 |
| `spec.connection.retryBackoffMs` | Minimum 100 |

## Watching PingoraConfig

The controller watches PingoraConfig resources and reconciles when:

- PingoraConfig is created, updated, or deleted
- Referenced Secret changes (if TLS enabled)
- GatewayClass parametersRef changes

## Next Steps

- Review [Helm Chart](helm-chart.md) for deployment options
- Check [Security](security.md) best practices
