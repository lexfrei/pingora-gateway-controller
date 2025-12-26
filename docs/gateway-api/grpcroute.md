# GRPCRoute

GRPCRoute defines gRPC routing rules for traffic entering through a Gateway.

## Basic Example

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: grpc-route
  namespace: default
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - grpc.example.com
  rules:
    - backendRefs:
        - name: grpc-service
          port: 50051
```

## Method Matching

### Service and Method

Route based on gRPC service and method:

```yaml
rules:
  - matches:
      - method:
          service: helloworld.Greeter
          method: SayHello
    backendRefs:
      - name: greeter-service
        port: 50051
```

### Service Only

Route all methods of a service:

```yaml
rules:
  - matches:
      - method:
          service: helloworld.Greeter
    backendRefs:
      - name: greeter-service
        port: 50051
```

### Regex Matching

Use regex patterns for flexible matching:

```yaml
rules:
  - matches:
      - method:
          type: RegularExpression
          service: ".*\\.v2\\..*"
    backendRefs:
      - name: v2-service
        port: 50051
```

## Header Matching

Route based on gRPC metadata (headers):

```yaml
rules:
  - matches:
      - headers:
          - name: x-tenant-id
            value: acme
    backendRefs:
      - name: acme-backend
        port: 50051
  - matches:
      - headers:
          - name: x-tenant-id
            value: globex
    backendRefs:
      - name: globex-backend
        port: 50051
```

### Combined Method and Header

```yaml
rules:
  - matches:
      - method:
          service: helloworld.Greeter
        headers:
          - name: x-environment
            value: staging
    backendRefs:
      - name: staging-greeter
        port: 50051
```

## Weighted Backends

Split gRPC traffic between services:

```yaml
rules:
  - matches:
      - method:
          service: payment.PaymentService
    backendRefs:
      - name: payment-v1
        port: 50051
        weight: 80
      - name: payment-v2
        port: 50051
        weight: 20
```

## Multiple Services

Route different services to different backends:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: multi-service-route
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - api.example.com
  rules:
    # User service
    - matches:
        - method:
            service: user.UserService
      backendRefs:
        - name: user-service
          port: 50051

    # Order service
    - matches:
        - method:
            service: order.OrderService
      backendRefs:
        - name: order-service
          port: 50051

    # Default: reflection and health
    - backendRefs:
        - name: default-grpc
          port: 50051
```

## Cross-Namespace Backend

Reference a gRPC service in another namespace:

```yaml
rules:
  - matches:
      - method:
          service: shared.AuthService
    backendRefs:
      - name: auth-service
        namespace: auth-system
        port: 50051
```

!!! note "ReferenceGrant Required"

    Cross-namespace references require a [ReferenceGrant](referencegrant.md)
    in the target namespace.

## Health Checking

Route gRPC health check requests:

```yaml
rules:
  # Health check service
  - matches:
      - method:
          service: grpc.health.v1.Health
    backendRefs:
      - name: health-service
        port: 50051

  # Main application
  - backendRefs:
      - name: app-service
        port: 50051
```

## Reflection Service

Route gRPC reflection requests:

```yaml
rules:
  # Reflection (for grpcurl, etc.)
  - matches:
      - method:
          service: grpc.reflection.v1alpha.ServerReflection
    backendRefs:
      - name: reflection-service
        port: 50051
```

## Complete Example

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: comprehensive-grpc-route
  namespace: default
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - grpc.example.com
  rules:
    # Health check with dedicated backend
    - matches:
        - method:
            service: grpc.health.v1.Health
      backendRefs:
        - name: health-service
          port: 50051

    # Canary deployment for new service version
    - matches:
        - method:
            service: user.UserService
        headers:
          - name: x-canary
            value: "true"
      backendRefs:
        - name: user-service-canary
          port: 50051

    # Production user service
    - matches:
        - method:
            service: user.UserService
      backendRefs:
        - name: user-service
          port: 50051
          weight: 95
        - name: user-service-canary
          port: 50051
          weight: 5

    # Order service
    - matches:
        - method:
            service: order.OrderService
      backendRefs:
        - name: order-service
          port: 50051

    # Default fallback
    - backendRefs:
        - name: default-grpc-backend
          port: 50051
```

## Testing with grpcurl

Test your GRPCRoute configuration:

```bash
# List services (requires reflection)
grpcurl -plaintext grpc.example.com:80 list

# Call a method
grpcurl -plaintext \
  -d '{"name": "World"}' \
  grpc.example.com:80 \
  helloworld.Greeter/SayHello

# With custom header
grpcurl -plaintext \
  -H 'x-tenant-id: acme' \
  -d '{}' \
  grpc.example.com:80 \
  user.UserService/GetUser
```

## Next Steps

- Set up [ReferenceGrant](referencegrant.md) for cross-namespace routing
- Review [HTTPRoute](httproute.md) for HTTP traffic
- Check [Limitations](limitations.md) for unsupported features
