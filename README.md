# Pingora Gateway Controller

[![Go Version](https://img.shields.io/github/go-mod/go-version/lexfrei/pingora-gateway-controller)](https://go.dev/)
[![License](https://img.shields.io/github/license/lexfrei/pingora-gateway-controller)](LICENSE)
[![CI](https://github.com/lexfrei/pingora-gateway-controller/actions/workflows/pr.yaml/badge.svg)](https://github.com/lexfrei/pingora-gateway-controller/actions/workflows/pr.yaml)

Kubernetes Gateway API controller for [Pingora](https://github.com/cloudflare/pingora) HTTP proxy framework.

Route traffic through Pingora using standard Gateway API resources (Gateway, HTTPRoute, GRPCRoute).

## Status

**Alpha** - This project is under active development targeting Gateway API v1.4.1 conformance.

| Milestone | Status | Description |
|-----------|--------|-------------|
| v0.1.0 | Planned | Core Conformance (RequestHeaderModifier, RequestRedirect) |
| v0.2.0 | Planned | Extended Features (ResponseHeaderModifier, URLRewrite, RequestMirror) |
| v0.3.0 | Planned | Gateway API v1.4 Features (supportedFeatures, BackendTLSPolicy) |
| v0.4.0 | Planned | Conformance Tests & Registration |
| v1.0.0 | Planned | Production-ready with full conformance |

## Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
├─────────────────────────────────────────────────────────────┤
│  HTTPRoute/GRPCRoute/Gateway                                │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────────────┐      gRPC      ┌───────────────┐  │
│  │ pingora-gateway-     │ ─────────────► │ pingora-proxy │  │
│  │ controller (Go)      │                │ (Rust)        │  │
│  └──────────────────────┘                └───────────────┘  │
│                                                 │           │
│                                                 ▼           │
│                                          HTTP/gRPC traffic  │
└─────────────────────────────────────────────────────────────┘
```

### Components

- **pingora-gateway-controller** (Go): Kubernetes controller that watches Gateway API resources and syncs routing configuration to Pingora proxy via gRPC
- **pingora-proxy** (Rust): Custom Pingora-based reverse proxy with gRPC API for dynamic route updates

## Gateway API Support

Target: **Gateway API v1.4.1 Standard Channel**

### Resources

| Resource | Status | Notes |
|----------|--------|-------|
| GatewayClass | Supported | with parametersRef to PingoraConfig |
| Gateway | Supported | listeners, statuses, attachedRoutes |
| HTTPRoute | Supported | matches, backendRefs, timeouts |
| GRPCRoute | Supported | service/method matching |
| ReferenceGrant | Supported | cross-namespace validation |
| BackendTLSPolicy | Planned | [#30](https://github.com/lexfrei/pingora-gateway-controller/issues/30) |

### HTTPRoute Features

| Feature | Conformance | Status |
|---------|-------------|--------|
| Path matching (Exact, Prefix, Regex) | Core | Supported |
| Header matching | Core | Supported |
| Query parameter matching | Extended | Supported |
| Method matching | Extended | Supported |
| Backend weight | Core | Supported |
| Request timeout | Extended | Supported |
| RequestHeaderModifier | Core | [Planned #23](https://github.com/lexfrei/pingora-gateway-controller/issues/23) |
| RequestRedirect | Core | [Planned #24](https://github.com/lexfrei/pingora-gateway-controller/issues/24) |
| ResponseHeaderModifier | Extended | [Planned #25](https://github.com/lexfrei/pingora-gateway-controller/issues/25) |
| URLRewrite | Extended | [Planned #26](https://github.com/lexfrei/pingora-gateway-controller/issues/26) |
| RequestMirror | Extended | [Planned #27](https://github.com/lexfrei/pingora-gateway-controller/issues/27) |

### GRPCRoute Features

| Feature | Conformance | Status |
|---------|-------------|--------|
| Service/Method matching | Core | Supported |
| Header matching | Core | Supported |
| Backend weight | Core | Supported |
| RequestHeaderModifier | Core | [Planned #23](https://github.com/lexfrei/pingora-gateway-controller/issues/23) |
| ResponseHeaderModifier | Extended | [Planned #25](https://github.com/lexfrei/pingora-gateway-controller/issues/25) |
| RequestMirror | Extended | [Planned #27](https://github.com/lexfrei/pingora-gateway-controller/issues/27) |

## Features

- Standard Gateway API implementation (GatewayClass, Gateway, HTTPRoute, GRPCRoute)
- Dynamic route updates via gRPC (no proxy restart required)
- Cross-namespace backend references with ReferenceGrant support
- Multi-arch container images (amd64, arm64)
- Prometheus metrics

## Prerequisites

- Kubernetes cluster (1.25+)
- Gateway API CRDs installed
- Pingora proxy deployed with gRPC API enabled

## Quick Start

```bash
# 1. Install Gateway API CRDs
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml

# 2. Install the controller (Helm chart coming soon)
# See docs for manual installation

# 3. Create HTTPRoute to expose your service
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-app
spec:
  parentRefs:
    - name: pingora
      namespace: pingora-system
  hostnames:
    - app.example.com
  rules:
    - backendRefs:
        - name: my-service
          port: 80
EOF
```

## Documentation

Full documentation is available at [pingora-gw.k8s.lex.la](https://pingora-gw.k8s.lex.la).

## Roadmap

See the [project milestones](https://github.com/lexfrei/pingora-gateway-controller/milestones) for planned features.

Key upcoming features:

- [#23](https://github.com/lexfrei/pingora-gateway-controller/issues/23) RequestHeaderModifier filter (Core)
- [#24](https://github.com/lexfrei/pingora-gateway-controller/issues/24) RequestRedirect filter (Core)
- [#25](https://github.com/lexfrei/pingora-gateway-controller/issues/25) ResponseHeaderModifier filter (Extended)
- [#26](https://github.com/lexfrei/pingora-gateway-controller/issues/26) URLRewrite filter (Extended)
- [#27](https://github.com/lexfrei/pingora-gateway-controller/issues/27) RequestMirror filter (Extended)
- [#30](https://github.com/lexfrei/pingora-gateway-controller/issues/30) BackendTLSPolicy support
- [#31](https://github.com/lexfrei/pingora-gateway-controller/issues/31) Gateway API conformance tests

## License

BSD 3-Clause License - see [LICENSE](LICENSE) for details.
