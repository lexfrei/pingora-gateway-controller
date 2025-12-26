# Testing

Testing patterns and practices for Pingora Gateway Controller.

## Testing Philosophy

- **TDD (Test-Driven Development)**: Write tests first, then implementation
- **Table-driven tests**: Use structured test cases
- **Parallel execution**: Tests should be parallelizable
- **Meaningful assertions**: Test behavior, not implementation

## Running Tests

### All Tests

```bash
go test -race ./...
```

### Specific Package

```bash
go test -v -race ./internal/controller/...
```

### Single Test

```bash
go test -v -race ./internal/controller/... -run TestHTTPRouteReconciler
```

### With Coverage

```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Short Tests Only

```bash
go test -short ./...
```

## Test Structure

### Package Layout

```text
internal/
├── controller/
│   ├── gateway_controller.go
│   ├── gateway_controller_test.go
│   ├── httproute_controller.go
│   └── httproute_controller_test.go
├── ingress/
│   ├── pingora_builder.go
│   └── pingora_builder_test.go
└── dns/
    ├── detect.go
    └── detect_test.go
```

### Test File Naming

- Test files: `*_test.go`
- Same package as code under test
- Internal tests access private members

## Test Patterns

### Table-Driven Tests

```go
func TestBuildHTTPRoute(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        route    *gatewayv1.HTTPRoute
        expected *routingv1.HTTPRoute
    }{
        {
            name: "basic route",
            route: &gatewayv1.HTTPRoute{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-route",
                    Namespace: "default",
                },
                Spec: gatewayv1.HTTPRouteSpec{
                    Hostnames: []gatewayv1.Hostname{"example.com"},
                },
            },
            expected: &routingv1.HTTPRoute{
                Id:        "default/test-route",
                Hostnames: []string{"example.com"},
            },
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            builder := NewPingoraBuilder("cluster.local")
            result := builder.BuildHTTPRoute(tt.route)

            assert.Equal(t, tt.expected.Id, result.Id)
            assert.Equal(t, tt.expected.Hostnames, result.Hostnames)
        })
    }
}
```

### Fake Client Setup

```go
func setupFakeClient(t *testing.T, objs ...client.Object) client.Client {
    t.Helper()

    scheme := runtime.NewScheme()
    require.NoError(t, clientgoscheme.AddToScheme(scheme))
    require.NoError(t, gatewayv1.AddToScheme(scheme))
    require.NoError(t, v1alpha1.AddToScheme(scheme))

    return fake.NewClientBuilder().
        WithScheme(scheme).
        WithObjects(objs...).
        Build()
}
```

### Controller Tests

```go
func TestHTTPRouteReconciler_Reconcile(t *testing.T) {
    t.Parallel()

    gateway := &gatewayv1.Gateway{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-gateway",
            Namespace: "pingora-system",
        },
        Spec: gatewayv1.GatewaySpec{
            GatewayClassName: "pingora",
        },
    }

    route := &gatewayv1.HTTPRoute{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-route",
            Namespace: "default",
        },
        Spec: gatewayv1.HTTPRouteSpec{
            ParentRefs: []gatewayv1.ParentReference{
                {
                    Name:      "test-gateway",
                    Namespace: ptr.To(gatewayv1.Namespace("pingora-system")),
                },
            },
        },
    }

    client := setupFakeClient(t, gateway, route)
    reconciler := &HTTPRouteReconciler{
        Client: client,
        Scheme: client.Scheme(),
    }

    req := ctrl.Request{
        NamespacedName: types.NamespacedName{
            Name:      "test-route",
            Namespace: "default",
        },
    }

    result, err := reconciler.Reconcile(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Requeue)
}
```

## Testing Libraries

### testify

Used for assertions and requirements:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// assert - continues on failure
assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.Len(t, slice, 3)

// require - stops on failure
require.NoError(t, err)
require.NotNil(t, result)
```

### Fake Client

controller-runtime fake client for unit tests:

```go
import "sigs.k8s.io/controller-runtime/pkg/client/fake"

client := fake.NewClientBuilder().
    WithScheme(scheme).
    WithObjects(objs...).
    WithStatusSubresource(statusObjs...).
    Build()
```

### envtest

Integration tests with real API server:

```go
import "sigs.k8s.io/controller-runtime/pkg/envtest"

func TestMain(m *testing.M) {
    testEnv := &envtest.Environment{
        CRDDirectoryPaths: []string{
            filepath.Join("..", "..", "config", "crd", "bases"),
        },
    }

    cfg, err := testEnv.Start()
    if err != nil {
        log.Fatal(err)
    }

    code := m.Run()

    testEnv.Stop()
    os.Exit(code)
}
```

## Helm Chart Tests

### helm-unittest

```yaml
# tests/deployment_test.yaml
suite: test deployment
templates:
  - templates/deployment.yaml
tests:
  - it: should create deployment
    asserts:
      - isKind:
          of: Deployment
      - equal:
          path: metadata.name
          value: RELEASE-NAME-pingora-gateway-controller
```

### Running Helm Tests

```bash
helm unittest charts/pingora-gateway-controller
```

## Mocking

### Interface-Based Mocking

```go
// Define interface
type Syncer interface {
    SyncRoutes(ctx context.Context, routes []*routingv1.HTTPRoute) error
}

// Mock implementation
type MockSyncer struct {
    SyncRoutesFunc func(ctx context.Context, routes []*routingv1.HTTPRoute) error
}

func (m *MockSyncer) SyncRoutes(ctx context.Context, routes []*routingv1.HTTPRoute) error {
    return m.SyncRoutesFunc(ctx, routes)
}

// Use in tests
syncer := &MockSyncer{
    SyncRoutesFunc: func(ctx context.Context, routes []*routingv1.HTTPRoute) error {
        assert.Len(t, routes, 1)
        return nil
    },
}
```

## Test Coverage

### Generating Coverage Report

```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Guidelines

- Aim for >80% coverage on business logic
- Focus on critical paths
- Don't chase 100% coverage
- Test error conditions

## CI Testing

Tests run automatically on:

- Pull request creation
- Push to feature branch
- Merge to master

### CI Checks

1. Unit tests: `go test -race ./...`
2. Linting: `golangci-lint run`
3. Helm tests: `helm unittest charts/...`
4. Documentation: `mkdocs build --strict`

## Debugging Tests

### Verbose Output

```bash
go test -v ./internal/controller/... -run TestHTTPRoute
```

### With Debug Logging

```go
func TestWithDebug(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))
    // Use logger in test
}
```

### Debugging Specific Test

```bash
# Run single test with verbose output
go test -v -race ./internal/controller/... -run TestHTTPRouteReconciler/case_name

# With race detector and CPU profiling
go test -v -race -cpuprofile=cpu.prof ./internal/controller/... -run TestHTTPRouteReconciler
```

## Next Steps

- Read [Contributing Guidelines](contributing.md)
- Understand the [Architecture](architecture.md)
