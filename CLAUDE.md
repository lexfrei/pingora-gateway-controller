# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kubernetes controller implementing Gateway API for Pingora proxy. Watches Gateway and HTTPRoute/GRPCRoute resources, syncs routing configuration to Pingora proxy via gRPC. Two-component architecture: Go controller + Rust Pingora proxy (in separate repo as git submodule).

## Build and Development Commands

```bash
# Build binary
go build -o bin/controller ./cmd/controller

# Build with version info
go build -ldflags "-X main.Version=v0.0.1 -X main.Gitsha=$(git rev-parse HEAD)" -o bin/controller ./cmd/controller

# Run tests
go test -v -race -coverprofile=coverage.out ./...

# Run single test
go test -v -race ./internal/dns/... -run TestDetectClusterDomain

# Run linter (all errors must be fixed before committing)
golangci-lint run --timeout=5m

# Lint markdown files
markdownlint-cli2 '**/*.md'

# Build container
podman build --tag pingora-gateway-controller:dev --file Containerfile .

# Generate protobuf (after schema is defined)
buf generate
```

## Helm Chart Commands

```bash
# Package chart
helm package charts/pingora-gateway-controller

# Run helm-unittest
helm unittest charts/pingora-gateway-controller

# Generate README from values.yaml (REQUIRED before commit)
helm-docs charts/pingora-gateway-controller

# Lint chart
helm lint charts/pingora-gateway-controller

# Template locally (for debugging)
helm template test charts/pingora-gateway-controller --values charts/pingora-gateway-controller/examples/basic-values.yaml
```

## Architecture

### Components

- **pingora-gateway-controller** (Go): Kubernetes controller that watches Gateway API resources and syncs routing configuration to Pingora proxy via gRPC
- **pingora-proxy** (Rust): Custom Pingora-based reverse proxy with gRPC API for dynamic route updates (separate repo, git submodule at `proxy/`)

### Controllers (controller-runtime based)

- **GatewayReconciler** (`internal/controller/gateway_controller.go`): Watches Gateway resources matching `pingora` GatewayClass. Updates Gateway status.

- **HTTPRouteReconciler** (`internal/controller/httproute_controller.go`): Watches HTTPRoute resources referencing managed Gateways. Syncs routes to Pingora via gRPC. Updates HTTPRoute status.

- **GRPCRouteReconciler** (`internal/controller/grpcroute_controller.go`): Watches GRPCRoute resources referencing managed Gateways. Syncs routes to Pingora via gRPC. Updates GRPCRoute status.

### Custom Resource Definition

- **PingoraConfig** (`api/v1alpha1/`): Cluster-scoped CRD for configuring Pingora proxy connection. Referenced by GatewayClass via `parametersRef`. Contains gRPC endpoint address and TLS configuration.

### Supporting Packages

- **internal/config/pingora_resolver.go**: Resolves PingoraConfig from GatewayClass parametersRef, creates gRPC client connection.

- **internal/controller/pingora_syncer.go**: Converts routes to protobuf format and calls Pingora gRPC API.

- **internal/ingress/pingora_builder.go**: Converts HTTPRoute/GRPCRoute specs to Pingora route format.

- **internal/dns/detect.go**: Auto-detects Kubernetes cluster domain from `/etc/resolv.conf` search domains.

- **pkg/api/routing/v1/**: Generated Go gRPC client from protobuf schema.

### Key Dependencies

- `sigs.k8s.io/controller-runtime` - Kubernetes controller framework
- `sigs.k8s.io/gateway-api` - Gateway API types
- `google.golang.org/grpc` - gRPC client
- `google.golang.org/protobuf` - Protocol Buffers
- `github.com/cockroachdb/errors` - Error wrapping

### Configuration

Configuration is provided via PingoraConfig CRD (referenced by GatewayClass parametersRef):

- `address` - Pingora proxy gRPC endpoint address (required)
- `tls.enabled` - Enable TLS for gRPC connection
- `tls.secretRef` - Secret with TLS certificates

## Project Structure

```text
api/
  proto/routing/v1/      # Protobuf schema for gRPC API
  v1alpha1/              # PingoraConfig CRD types
cmd/controller/          # Entrypoint and CLI (cobra/viper)
internal/
  config/                # PingoraConfig resolver and gRPC client setup
  controller/            # Kubernetes controllers (Gateway, HTTPRoute, GRPCRoute)
  dns/                   # Cluster domain auto-detection
  ingress/               # Route → Pingora format conversion
  metrics/               # Prometheus metrics
pkg/api/routing/v1/      # Generated Go gRPC client
proxy/                   # Git submodule: pingora-proxy (Rust)
charts/                  # Helm chart with helm-unittest tests
deploy/                  # Raw Kubernetes manifests for manual deployment
docs/                    # MkDocs documentation
```

## Testing Standards

### Approach

- **TDD (Test-Driven Development)**: Write tests first, then implementation
- Follow RED → GREEN → REFACTOR cycle
- Commit test and implementation together per feature

### Testing Libraries

- `github.com/stretchr/testify/assert` - Assertions
- `github.com/stretchr/testify/require` - Fatal assertions (stops test on failure)
- `sigs.k8s.io/controller-runtime/pkg/client/fake` - Fake Kubernetes client for unit tests
- `sigs.k8s.io/controller-runtime/pkg/envtest` - Integration tests with real API server

### Test Patterns

- **Table-driven tests**: Use `[]struct{}` with named test cases
- **Parallel execution**: Always use `t.Parallel()` at test and subtest level
- **Fake client setup**: Create scheme, register types, build fake client
- **Helper functions**: Extract common setup (e.g., `setupFakeClient()`)

### Example Structure

```go
func TestFeature(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        input    InputType
        expected OutputType
    }{
        {name: "case 1", input: ..., expected: ...},
        {name: "case 2", input: ..., expected: ...},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // test logic
            require.NoError(t, err)
            assert.Equal(t, tt.expected, actual)
        })
    }
}
```

### Running Tests

```bash
# All tests with race detection
go test -race ./...

# Single package
go test -v -race ./internal/controller/...

# Single test by name
go test -v -race ./internal/controller/... -run TestHTTPRouteReconciler

# With coverage
go test -race -coverprofile=coverage.out ./...
```

## Documentation

Project documentation uses MkDocs with Material theme. Live site: https://pingora.k8s.lex.la

### Commands

```bash
# Install dependencies
pip install --requirement requirements-docs.txt

# Local preview server
mkdocs serve

# Build static site
mkdocs build

# Lint markdown
markdownlint-cli2 '**/*.md'
```

### Structure

```text
docs/
├── index.md                 # Homepage
├── getting-started/         # Installation, prerequisites, quickstart
├── configuration/           # Controller options, Helm values, PingoraConfig
├── gateway-api/             # HTTPRoute, GRPCRoute, ReferenceGrant, limitations
├── guides/                  # Cross-namespace routing, monitoring
├── operations/              # Troubleshooting, metrics, manual installation
├── development/             # Setup, architecture, contributing, testing
└── reference/               # Helm chart, CRD reference, security
```

### Writing Guidelines

- **Section structure**: Each section must have `index.md` as landing page
- **Navigation**: Register all new pages in `nav:` section of `mkdocs.yml`
- **Diagrams**: Use Mermaid for architecture and flow diagrams
- **Code blocks**: Use syntax highlighting with language identifier
- **Admonitions**: Use `!!! note`, `!!! warning`, `!!! danger` for callouts
- **Links**: Use relative paths for internal links (`../configuration/helm-values.md`)

### Documentation TDD

**CRITICAL: Apply TDD methodology to documentation with obsessive attention to detail.**

**NEVER work directly on master branch. Create a feature branch for all documentation changes.**

Before writing any documentation:

1. **Verify every command works**
   - Run each command yourself before documenting
   - Test on clean environment when possible
   - Document exact versions and prerequisites

2. **Validate all code examples**
   - Copy-paste and execute every code snippet
   - Verify output matches documented expectations
   - Test edge cases mentioned in documentation

3. **Check all links and references**
   - Click every internal link
   - Verify external URLs are accessible
   - Confirm file paths exist

4. **Test the user journey**
   - Follow your own documentation step-by-step
   - Assume zero prior knowledge
   - Note every missing step or assumption

5. **Build and preview locally**
   - Run `mkdocs serve` before committing
   - Check rendering of all changed pages
   - Verify navigation and search work correctly

**If a command fails, a link is broken, or a step is missing — fix it before committing.**

## Linting Configuration

golangci-lint v2 config in `.golangci.yaml`:

- `funlen` limit: 60 lines/statements
- `gocyclo/cyclop` complexity: 15
- All linters enabled by default with specific exclusions
- Test files have relaxed rules for funlen, dupl, complexity

## Pull Request Guidelines

### Local CI Checks

**CRITICAL: Run CI-equivalent checks locally before each push, not just before PR creation.**

Run checks relevant to the files you changed:

| Changed Files | Required Checks |
|---------------|-----------------|
| `*.go` | `go test -race ./...` and `golangci-lint run --timeout=5m` |
| `charts/**` | `helm unittest`, `helm lint`, `helm-docs` |
| `**/*.md` | `markdownlint-cli2 '**/*.md'` |
| `docs/**` | `mkdocs build --strict` |

Quick reference commands:

```bash
# Go code changes
go test -race ./... && golangci-lint run --timeout=5m

# Helm chart changes
helm unittest charts/pingora-gateway-controller && \
helm lint charts/pingora-gateway-controller && \
helm-docs charts/pingora-gateway-controller

# Markdown changes
markdownlint-cli2 '**/*.md'

# Documentation site
mkdocs build --strict
```

**Why this matters:** CI failures after push waste time, trigger unnecessary notifications, and delay reviews. Catching issues locally is faster and cheaper.

### Pre-PR Checklist

Before creating a PR, verify all checklist items from `.github/pull_request_template.md`:

1. **Testing**
   - All tests pass locally (`go test ./...`)
   - Linters pass locally (`golangci-lint run`)
   - Markdown linting passes (`markdownlint-cli2 '**/*.md'`)
   - Helm tests pass (`helm unittest charts/pingora-gateway-controller`)
   - Helm lint passes (`helm lint charts/pingora-gateway-controller`)
   - Helm README is up to date (`helm-docs charts/pingora-gateway-controller`)
   - Manual testing completed (if applicable)

2. **Documentation**
   - README updated (if needed)
   - Code comments added for complex logic
   - CLAUDE.md updated (if workflow/standards changed)

3. **Code Quality**
   - Commit messages follow semantic format (`type(scope): description`)
   - No secrets or credentials in code
   - Breaking changes documented (if any)

### PR Creation

- Use template from `.github/pull_request_template.md`
- Fill all sections completely
- Check all applicable checkboxes honestly
- Do NOT check boxes for items not actually completed

## GitHub Issue Labels

When creating issues, apply labels from these categories:

### Type (required)

- `bug` — Something isn't working
- `enhancement` — New feature or request
- `documentation` — Documentation improvements
- `test` — Test coverage
- `ci` — CI/CD and automation
- `security` — Security-related

### Area (required)

- `area/controller` — Controller code
- `area/helm` — Helm chart
- `area/api` — CRD and API types
- `area/docs` — Documentation
- `area/proxy` — Pingora proxy (Rust)
- `area/grpc` — gRPC API and protobuf

### Priority (required)

- `priority/critical` — Blocks release, needs immediate attention
- `priority/high` — Important for milestone
- `priority/medium` — Should be done for milestone
- `priority/low` — Nice to have, can defer

### Status (required)

- `status/needs-triage` — Requires analysis
- `status/needs-design` — Requires design/RFC
- `status/ready` — Ready to work on
- `status/in-progress` — Currently being worked on
- `status/blocked` — Blocked by dependency
- `status/needs-info` — Waiting for clarification
- `status/needs-review` — Waiting for review/feedback

### Size (required)

- `size/XS` — < 1 hour
- `size/S` — 1-4 hours
- `size/M` — 1-2 days
- `size/L` — 3-5 days
- `size/XL` — > 1 week

### Milestone

Always assign a milestone when creating issues (e.g., `v0.0.1`).
