# Limitations

Current limitations and unsupported features in Pingora Gateway Controller.

## Unsupported Features

### HTTPRoute Filters

The following HTTPRoute filters are not currently supported:

| Filter | Status | Alternative |
|--------|--------|-------------|
| RequestHeaderModifier | Not Supported | Backend handling |
| ResponseHeaderModifier | Not Supported | Backend handling |
| RequestRedirect | Not Supported | Backend handling |
| URLRewrite | Not Supported | Backend handling |
| RequestMirror | Not Supported | - |
| ExtensionRef | Not Supported | - |

!!! note "Future Support"

    Filter support is planned for future releases. Track progress in
    [GitHub Issues](https://github.com/lexfrei/pingora-gateway-controller/issues).

### TLS Configuration

| Feature | Status | Notes |
|---------|--------|-------|
| TLS termination | Planned | Future release |
| TLS passthrough | Not Planned | Backend handles TLS |
| mTLS | Planned | Client certificate validation |
| Certificate rotation | Planned | Automatic Secret reload |

### Backend Types

Only Kubernetes Service backends are supported:

| Backend Type | Status |
|--------------|--------|
| Service | Supported |
| External addresses | Not Supported |
| Custom backends | Not Supported |

### Gateway Features

| Feature | Status | Notes |
|---------|--------|-------|
| Multiple Gateways | Supported | Same GatewayClass |
| Gateway merge | Not Supported | Use single Gateway |
| Infrastructure | Not Supported | - |

### Listener Features

| Feature | Status | Notes |
|---------|--------|-------|
| HTTP | Supported | Default |
| HTTPS | Planned | TLS termination |
| TLS | Planned | Passthrough |
| Allowed routes | Partial | Namespace selector only |

## Known Issues

### Route Priority

When multiple routes match a request, the order of evaluation follows
Gateway API specification:

1. Exact path matches before prefix matches
2. Longer prefixes before shorter prefixes
3. Header matches add specificity
4. Method matches add specificity

!!! warning "Regex Priority"

    Regex path matches have lower priority than exact and prefix matches.
    This may differ from some other implementations.

### Backend Resolution

- Backend services must have at least one ready endpoint
- DNS resolution uses cluster domain (auto-detected or configured)
- Service ports must be explicitly specified

### Cross-Namespace References

- ReferenceGrant is required for all cross-namespace references
- Controller does not cache ReferenceGrant - changes apply immediately
- Deleting ReferenceGrant immediately breaks affected routes

## Resource Limits

| Resource | Limit | Notes |
|----------|-------|-------|
| Routes per Gateway | No hard limit | Performance degrades with thousands |
| Rules per Route | No hard limit | Keep reasonable for maintainability |
| Backends per Rule | No hard limit | - |
| Hostnames per Route | No hard limit | - |

## Performance Considerations

### Route Sync

- Routes are synced to Pingora proxy via gRPC
- Large configurations may take longer to sync
- Consider splitting routes across multiple HTTPRoute resources

### Regex Matching

- Regex patterns are compiled at sync time
- Complex patterns impact matching performance
- Use exact or prefix matching when possible

### Memory Usage

- Controller memory scales with number of watched resources
- Proxy memory scales with active connections and route count
- Monitor resource usage in production

## Workarounds

### Header Modification

Use a sidecar or middleware in your application:

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: app
          # Your application
        - name: header-modifier
          image: envoyproxy/envoy:v1.28.0
          # Configure Envoy for header modification
```

### URL Rewriting

Handle URL rewriting in your application code or use a reverse proxy
sidecar.

### Request Mirroring

Use service mesh features (Istio, Linkerd) if request mirroring is
required.

## Roadmap

Features planned for future releases:

1. **TLS Support** - HTTPS listeners with certificate management
2. **Filters** - Request/response modification
3. **Policy Attachment** - Gateway API Policy resources
4. **TCPRoute/UDPRoute** - Layer 4 routing

Track development progress on [GitHub](https://github.com/lexfrei/pingora-gateway-controller).

## Reporting Issues

If you encounter issues or need features not listed here:

1. Check [existing issues](https://github.com/lexfrei/pingora-gateway-controller/issues)
2. Create a new issue with:
   - Clear description of the problem
   - Expected vs actual behavior
   - Minimal reproduction steps
   - Controller and Gateway API versions

## Next Steps

- Review [Supported Resources](supported-resources.md) for current capabilities
- Check [Troubleshooting](../operations/troubleshooting.md) for common issues
