# Development Setup

Set up your local development environment for Pingora Gateway Controller.

## Prerequisites

### Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.23+ | Build and test |
| kubectl | Latest | Kubernetes CLI |
| Helm | 3.x | Chart testing |
| golangci-lint | Latest | Code linting |
| buf | Latest | Protobuf generation |

### Optional Tools

| Tool | Purpose |
|------|---------|
| kind | Local Kubernetes cluster |
| podman/docker | Container building |
| helm-unittest | Helm chart testing |
| markdownlint-cli2 | Documentation linting |

## Installation

### macOS (Homebrew)

```bash
# Go
brew install go

# Kubernetes tools
brew install kubectl helm

# Development tools
brew install golangci-lint buf

# Optional
brew install kind podman helm-unittest markdownlint-cli2
```

### Linux

```bash
# Go (from go.dev)
wget https://go.dev/dl/go1.23.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# buf
GO111MODULE=on go install github.com/bufbuild/buf/cmd/buf@latest
```

## Clone Repository

```bash
git clone https://github.com/lexfrei/pingora-gateway-controller.git
cd pingora-gateway-controller

# Initialize submodules (for proxy)
git submodule update --init --recursive
```

## Build

### Build Controller Binary

```bash
go build -o bin/controller ./cmd/controller
```

### Build with Version Info

```bash
go build -ldflags "-X main.Version=v0.0.1 -X main.Gitsha=$(git rev-parse HEAD)" \
  -o bin/controller ./cmd/controller
```

### Build Container Image

```bash
podman build --tag pingora-gateway-controller:dev --file Containerfile .
```

## Run Locally

### Against Remote Cluster

```bash
# Ensure kubectl is configured
kubectl config current-context

# Run controller
./bin/controller \
  --gateway-class-name=pingora \
  --log-level=debug \
  --log-format=text
```

### With kind Cluster

```bash
# Create cluster
kind create cluster --name pingora-dev

# Install Gateway API CRDs
kubectl apply --filename https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml

# Load local image
kind load docker-image pingora-gateway-controller:dev --name pingora-dev

# Install with Helm
helm install pingora-gateway-controller charts/pingora-gateway-controller \
  --namespace pingora-system \
  --create-namespace \
  --set image.tag=dev \
  --set image.pullPolicy=Never
```

## IDE Setup

### VS Code

Recommended extensions:

- Go (golang.go)
- YAML (redhat.vscode-yaml)
- Markdown All in One (yzhang.markdown-all-in-one)
- Kubernetes (ms-kubernetes-tools.vscode-kubernetes-tools)

Settings (`.vscode/settings.json`):

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

### GoLand/IntelliJ

- Enable golangci-lint integration
- Configure Go modules
- Set up Kubernetes plugin

## Development Workflow

### Make Changes

1. Create feature branch
2. Make changes
3. Run tests and linter
4. Commit with semantic message

### Test Cycle

```bash
# Run all tests
go test -race ./...

# Run specific package
go test -v -race ./internal/controller/...

# Run with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Lint Cycle

```bash
# Run linter
golangci-lint run

# Auto-fix where possible
golangci-lint run --fix
```

### Protobuf Generation

If modifying protobuf schema:

```bash
buf generate
```

## Helm Chart Development

### Run Tests

```bash
helm unittest charts/pingora-gateway-controller
```

### Generate README

```bash
helm-docs charts/pingora-gateway-controller
```

### Lint Chart

```bash
helm lint charts/pingora-gateway-controller
```

## Documentation Development

### Local Preview

```bash
# Install dependencies
pip install --requirement requirements-docs.txt

# Serve locally
mkdocs serve
```

### Build Documentation

```bash
mkdocs build --strict
```

### Lint Markdown

```bash
markdownlint-cli2 '**/*.md'
```

## Troubleshooting

### Go Module Issues

```bash
go mod tidy
go mod download
```

### Submodule Issues

```bash
git submodule update --init --recursive
```

### golangci-lint Cache

```bash
golangci-lint cache clean
```

## Next Steps

- Understand the [Architecture](architecture.md)
- Read [Contributing Guidelines](contributing.md)
- Learn about [Testing](testing.md)
