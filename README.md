# Pingora Gateway Controller

[![Go Version](https://img.shields.io/github/go-mod/go-version/lexfrei/pingora-gateway-controller)](https://go.dev/)
[![License](https://img.shields.io/github/license/lexfrei/pingora-gateway-controller)](LICENSE)
[![CI](https://github.com/lexfrei/pingora-gateway-controller/actions/workflows/pr.yaml/badge.svg)](https://github.com/lexfrei/pingora-gateway-controller/actions/workflows/pr.yaml)

Kubernetes Gateway API controller for [Pingora](https://github.com/cloudflare/pingora) HTTP proxy framework.

Route traffic through Pingora using standard Gateway API resources (Gateway, HTTPRoute, GRPCRoute).

## Status

**Work in Progress** - This project is under active development.

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
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml

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

## License

BSD 3-Clause License - see [LICENSE](LICENSE) for details.
