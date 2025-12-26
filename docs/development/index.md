# Development

This section covers development setup, architecture, and contribution guidelines.

## Sections

<div class="grid cards" markdown>

-   :material-laptop:{ .lg .middle } **Setup**

    ---

    Set up your local development environment with all required tools.

    [:octicons-arrow-right-24: Setup](setup.md)

-   :material-sitemap:{ .lg .middle } **Architecture**

    ---

    Deep dive into the controller architecture, components, and data flow.

    [:octicons-arrow-right-24: Architecture](architecture.md)

-   :material-source-branch:{ .lg .middle } **Contributing**

    ---

    Guidelines for contributing code, documentation, and reporting issues.

    [:octicons-arrow-right-24: Contributing](contributing.md)

-   :material-test-tube:{ .lg .middle } **Testing**

    ---

    Testing patterns, running tests, and writing new test cases.

    [:octicons-arrow-right-24: Testing](testing.md)

</div>

## Quick Start

```bash
# Clone repository
git clone https://github.com/lexfrei/pingora-gateway-controller.git
cd pingora-gateway-controller

# Install dependencies
go mod download

# Run tests
go test -race ./...

# Build binary
go build -o bin/controller ./cmd/controller

# Run linter
golangci-lint run
```

## Project Structure

```text
├── api/
│   ├── proto/routing/v1/    # Protobuf schema for gRPC API
│   └── v1alpha1/            # PingoraConfig CRD types
├── cmd/controller/          # Entrypoint and CLI
├── internal/
│   ├── config/              # PingoraConfig resolver
│   ├── controller/          # Kubernetes controllers
│   ├── dns/                 # Cluster domain detection
│   ├── ingress/             # Route → Pingora conversion
│   └── metrics/             # Prometheus metrics
├── pkg/api/routing/v1/      # Generated gRPC client
├── proxy/                   # Git submodule: Pingora proxy
├── charts/                  # Helm chart
├── deploy/                  # Raw Kubernetes manifests
└── docs/                    # Documentation
```

## Next Steps

- Set up [Development Environment](setup.md)
- Understand the [Architecture](architecture.md)
- Read [Contributing Guidelines](contributing.md)
