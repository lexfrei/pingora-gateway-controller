# Reference

Complete reference documentation for Pingora Gateway Controller.

## Sections

<div class="grid cards" markdown>

-   :material-file-document:{ .lg .middle } **Helm Chart**

    ---

    Complete Helm chart reference with all available values and options.

    [:octicons-arrow-right-24: Helm Chart](helm-chart.md)

-   :material-code-json:{ .lg .middle } **CRD Reference**

    ---

    Complete specification of the PingoraConfig Custom Resource Definition.

    [:octicons-arrow-right-24: CRD Reference](crd-reference.md)

-   :material-shield-lock:{ .lg .middle } **Security**

    ---

    Security considerations, best practices, and hardening guidelines.

    [:octicons-arrow-right-24: Security](security.md)

</div>

## Quick Links

### API Versions

| Resource | API Version |
|----------|-------------|
| PingoraConfig | `pingora.k8s.lex.la/v1alpha1` |
| GatewayClass | `gateway.networking.k8s.io/v1` |
| Gateway | `gateway.networking.k8s.io/v1` |
| HTTPRoute | `gateway.networking.k8s.io/v1` |
| GRPCRoute | `gateway.networking.k8s.io/v1` |

### Container Images

| Component | Image |
|-----------|-------|
| Controller | `ghcr.io/lexfrei/pingora-gateway-controller` |
| Proxy | `ghcr.io/lexfrei/pingora-proxy` |

### Default Ports

| Port | Purpose |
|------|---------|
| 80 | HTTP traffic (proxy) |
| 50051 | gRPC (proxy) |
| 8080 | Metrics (controller) |
| 8081 | Health checks |

## Next Steps

- Review [Helm Chart](helm-chart.md) configuration
- Check [CRD Reference](crd-reference.md) for custom resources
- Read [Security](security.md) best practices
