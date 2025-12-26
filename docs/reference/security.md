# Security

Security considerations and best practices for Pingora Gateway Controller.

## Container Security

### Default Security Context

The Helm chart applies secure defaults:

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65534
  seccompProfile:
    type: RuntimeDefault

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL
```

### Pod Security Standards

The controller is compatible with:

- **Restricted** Pod Security Standard (most secure)
- **Baseline** Pod Security Standard
- **Privileged** Pod Security Standard

To enforce restricted mode:

```bash
kubectl label namespace pingora-system \
  pod-security.kubernetes.io/enforce=restricted
```

## RBAC

### Controller Permissions

The controller requires minimal RBAC permissions:

| Resource | Verbs | Purpose |
|----------|-------|---------|
| GatewayClass | get, list, watch | Watch GatewayClass |
| GatewayClass/status | update, patch | Update status |
| Gateway | get, list, watch | Watch Gateway |
| Gateway/status | update, patch | Update status |
| HTTPRoute | get, list, watch | Watch HTTPRoute |
| HTTPRoute/status | update, patch | Update status |
| GRPCRoute | get, list, watch | Watch GRPCRoute |
| GRPCRoute/status | update, patch | Update status |
| ReferenceGrant | get, list, watch | Cross-namespace refs |
| PingoraConfig | get, list, watch | Configuration |
| PingoraConfig/status | update, patch | Update status |
| Service | get, list, watch | Backend resolution |
| Endpoints | get, list, watch | Backend health |
| Secret | get, list, watch | TLS certificates |
| Event | create, patch | Event recording |
| Lease | get, create, update | Leader election |

### Limiting Permissions

For additional security, consider:

1. **Namespace-scoped**: Configure controller to watch specific namespaces
2. **Custom ClusterRole**: Remove unused permissions
3. **Audit logging**: Enable Kubernetes audit logs

## Network Security

### Network Policies

Enable NetworkPolicy in Helm values:

```yaml
networkPolicy:
  enabled: true
  ingress:
    from:
      - namespaceSelector:
          matchLabels:
            name: monitoring
  pingoraProxy:
    port: 50051
    namespaceSelector:
      matchLabels:
        name: pingora-system
    podSelector:
      matchLabels:
        app.kubernetes.io/component: proxy
```

### Minimal Network Access

The controller needs:

| Direction | Target | Port | Purpose |
|-----------|--------|------|---------|
| Egress | Kubernetes API | 443 | Watch resources |
| Egress | Pingora Proxy | 50051 | gRPC sync |
| Ingress | Prometheus | 8080 | Metrics scraping |
| Ingress | Kubelet | 8081 | Health probes |

## TLS Configuration

### Controller to Proxy Communication

Enable TLS for gRPC communication:

```yaml
pingoraConfig:
  tls:
    enabled: true
    secretRef:
      name: controller-mtls
      namespace: pingora-system
```

### Certificate Requirements

- Use certificates from a trusted CA
- Enable mutual TLS (mTLS) for production
- Rotate certificates regularly
- Monitor certificate expiration

### Secret Management

For production, use external secret management:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: controller-mtls
  namespace: pingora-system
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault
    kind: ClusterSecretStore
  target:
    name: controller-mtls
  data:
    - secretKey: tls.crt
      remoteRef:
        key: pingora/controller/cert
    - secretKey: tls.key
      remoteRef:
        key: pingora/controller/key
    - secretKey: ca.crt
      remoteRef:
        key: pingora/ca/cert
```

## Image Security

### Image Verification

Verify container images before deployment:

```bash
# Check image digest
podman pull ghcr.io/lexfrei/pingora-gateway-controller:latest
podman inspect --format='{{.Digest}}' ghcr.io/lexfrei/pingora-gateway-controller:latest
```

### Image Pinning

Pin images to specific digests:

```yaml
image:
  repository: ghcr.io/lexfrei/pingora-gateway-controller
  tag: v0.1.0@sha256:abc123...
```

### Vulnerability Scanning

Scan images regularly:

```bash
trivy image ghcr.io/lexfrei/pingora-gateway-controller:latest
```

## Secrets Management

### Sensitive Data

The controller may access:

- TLS certificates for proxy communication
- Kubernetes API server credentials (via ServiceAccount)

### Best Practices

1. **Limit Secret access**: Only grant access to required Secrets
2. **Encrypt at rest**: Enable Kubernetes Secret encryption
3. **Audit access**: Monitor Secret access patterns
4. **Rotate regularly**: Implement certificate/token rotation

## Gateway API Security

### ReferenceGrant

Use ReferenceGrant to control cross-namespace access:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata:
  name: allow-specific-routes
  namespace: backend
spec:
  from:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      namespace: frontend  # Only frontend namespace
  to:
    - group: ""
      kind: Service
```

### Route Isolation

Consider namespace-based isolation:

- Separate Gateway per tenant
- Strict ReferenceGrant policies
- NetworkPolicy per namespace

## Monitoring and Alerting

### Security-Related Alerts

Monitor for:

```yaml
# Unexpected permission errors
- alert: PingoraRBACErrors
  expr: increase(controller_runtime_reconcile_total{result="error"}[5m]) > 10
  annotations:
    summary: "Possible RBAC misconfiguration"

# TLS connection failures
- alert: PingoraTLSErrors
  expr: increase(pingora_grpc_errors_total{error_type="tls"}[5m]) > 0
  annotations:
    summary: "TLS connection errors"
```

### Audit Logging

Enable Kubernetes audit logging for:

- Secret access
- ReferenceGrant changes
- Gateway/Route modifications

## Hardening Checklist

- [ ] Run as non-root user (default)
- [ ] Enable seccomp profile (default)
- [ ] Drop all capabilities (default)
- [ ] Read-only filesystem (default)
- [ ] Enable NetworkPolicy
- [ ] Enable TLS for gRPC
- [ ] Pin container images
- [ ] Scan images for vulnerabilities
- [ ] Enable Kubernetes audit logging
- [ ] Use external secret management
- [ ] Apply restrictive RBAC
- [ ] Monitor security events

## Reporting Security Issues

Report security vulnerabilities to:

- Email: security@lex.la
- GitHub Security Advisories

Do not report security issues in public GitHub issues.

## Next Steps

- Review [Helm Chart](helm-chart.md) security settings
- Check [CRD Reference](crd-reference.md) for TLS configuration
