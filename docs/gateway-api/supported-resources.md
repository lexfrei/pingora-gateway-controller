# Supported Resources

This page documents Gateway API resource support in Pingora Gateway Controller.

## Core Resources

| Resource | Status | Notes |
|----------|--------|-------|
| GatewayClass | Supported | Single GatewayClass per controller |
| Gateway | Supported | Multiple listeners supported |
| HTTPRoute | Supported | Full match support |
| GRPCRoute | Supported | Service/method matching |
| ReferenceGrant | Supported | Cross-namespace references |

## HTTPRoute Features

### Match Types

| Match Type | Status | Notes |
|------------|--------|-------|
| Path (Exact) | Supported | Exact path matching |
| Path (PathPrefix) | Supported | Prefix matching |
| Path (RegularExpression) | Supported | Regex patterns |
| Headers (Exact) | Supported | Exact value match |
| Headers (RegularExpression) | Supported | Regex patterns |
| QueryParams (Exact) | Supported | Exact value match |
| QueryParams (RegularExpression) | Supported | Regex patterns |
| Method | Supported | HTTP method filtering |

### Backend Features

| Feature | Status | Notes |
|---------|--------|-------|
| Service backends | Supported | Kubernetes Service only |
| Cross-namespace backends | Supported | With ReferenceGrant |
| Weighted backends | Supported | Traffic splitting |
| Port specification | Supported | Required for Service |

### Route Features

| Feature | Status | Notes |
|---------|--------|-------|
| Multiple hostnames | Supported | Per-route hostnames |
| Multiple rules | Supported | Ordered rule evaluation |
| Request timeouts | Supported | Per-rule timeout |
| Filters | Not Supported | See [Limitations](limitations.md) |

## GRPCRoute Features

### Match Types

| Match Type | Status | Notes |
|------------|--------|-------|
| Method (Exact) | Supported | Service + method name |
| Method (RegularExpression) | Supported | Regex patterns |
| Headers (Exact) | Supported | Exact value match |
| Headers (RegularExpression) | Supported | Regex patterns |

### Backend Features

| Feature | Status | Notes |
|---------|--------|-------|
| Service backends | Supported | Kubernetes Service only |
| Cross-namespace backends | Supported | With ReferenceGrant |
| Weighted backends | Supported | Traffic splitting |

## Gateway Features

### Listeners

| Feature | Status | Notes |
|---------|--------|-------|
| HTTP protocol | Supported | Default listener |
| HTTPS protocol | Planned | TLS termination |
| gRPC protocol | Supported | Via GRPCRoute |
| Multiple listeners | Supported | Same or different ports |

### TLS Configuration

| Feature | Status | Notes |
|---------|--------|-------|
| TLS termination | Planned | Future release |
| TLS passthrough | Not Planned | Backend handles TLS |
| Certificate rotation | Planned | Secret watch |

## ReferenceGrant

| Feature | Status | Notes |
|---------|--------|-------|
| Service references | Supported | Cross-namespace backends |
| Secret references | Planned | For TLS configuration |
| Gateway references | Supported | Cross-namespace parentRef |

## Status Updates

The controller updates resource status to reflect configuration state:

### Gateway Status

```yaml
status:
  conditions:
    - type: Accepted
      status: "True"
    - type: Programmed
      status: "True"
  listeners:
    - name: http
      conditions:
        - type: Accepted
          status: "True"
```

### HTTPRoute Status

```yaml
status:
  parents:
    - parentRef:
        name: pingora-gateway
        namespace: pingora-system
      conditions:
        - type: Accepted
          status: "True"
        - type: ResolvedRefs
          status: "True"
```

## Version Compatibility

| Gateway API Version | Controller Version | Status |
|--------------------|--------------------|--------|
| v1.4.1 | v0.x.x | Supported |
| v1.3.x | v0.x.x | Supported |
| v1.2.x | v0.x.x | Partial |
| v1.1.x | v0.x.x | Partial |
| < v1.1.0 | - | Not Supported |

!!! tip "Recommended Version"

    Use Gateway API v1.4.1 for full feature support.

## Next Steps

- Configure [HTTPRoute](httproute.md) for HTTP routing
- Configure [GRPCRoute](grpcroute.md) for gRPC services
- Review [Limitations](limitations.md) for unsupported features
