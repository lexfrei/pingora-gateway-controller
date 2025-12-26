# HTTPRoute

HTTPRoute defines HTTP routing rules for traffic entering through a Gateway.

## Basic Example

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: simple-route
  namespace: default
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - app.example.com
  rules:
    - backendRefs:
        - name: my-service
          port: 80
```

## Path Matching

### PathPrefix

Match requests starting with a path prefix:

```yaml
rules:
  - matches:
      - path:
          type: PathPrefix
          value: /api
    backendRefs:
      - name: api-service
        port: 8080
```

### Exact Path

Match requests with exact path:

```yaml
rules:
  - matches:
      - path:
          type: Exact
          value: /health
    backendRefs:
      - name: health-service
        port: 8080
```

### Regular Expression

Match paths using regex patterns:

```yaml
rules:
  - matches:
      - path:
          type: RegularExpression
          value: "/users/[0-9]+"
    backendRefs:
      - name: user-service
        port: 8080
```

!!! warning "Regex Performance"

    Complex regex patterns can impact routing performance. Use PathPrefix
    or Exact matching when possible.

## Header Matching

Route based on HTTP headers:

```yaml
rules:
  - matches:
      - headers:
          - name: X-Environment
            value: staging
    backendRefs:
      - name: staging-service
        port: 8080
  - matches:
      - headers:
          - name: X-Environment
            value: production
    backendRefs:
      - name: production-service
        port: 8080
```

### Regex Header Matching

```yaml
rules:
  - matches:
      - headers:
          - name: X-Request-ID
            type: RegularExpression
            value: "^[a-f0-9]{32}$"
    backendRefs:
      - name: traced-service
        port: 8080
```

## Method Matching

Route based on HTTP method:

```yaml
rules:
  - matches:
      - method: GET
        path:
          type: PathPrefix
          value: /api/v1
    backendRefs:
      - name: read-api
        port: 8080
  - matches:
      - method: POST
        path:
          type: PathPrefix
          value: /api/v1
    backendRefs:
      - name: write-api
        port: 8080
```

## Query Parameter Matching

Route based on query parameters:

```yaml
rules:
  - matches:
      - queryParams:
          - name: version
            value: v2
    backendRefs:
      - name: api-v2
        port: 8080
  - backendRefs:
      - name: api-v1
        port: 8080
```

### Regex Query Matching

```yaml
rules:
  - matches:
      - queryParams:
          - name: id
            type: RegularExpression
            value: "^[0-9]+$"
    backendRefs:
      - name: numeric-id-service
        port: 8080
```

## Multiple Match Conditions

Combine multiple match conditions (AND logic within a match):

```yaml
rules:
  - matches:
      - path:
          type: PathPrefix
          value: /api
        headers:
          - name: X-API-Version
            value: "2"
        method: POST
    backendRefs:
      - name: api-v2-write
        port: 8080
```

## Multiple Matches (OR logic)

Multiple matches within a rule use OR logic:

```yaml
rules:
  - matches:
      - path:
          type: Exact
          value: /health
      - path:
          type: Exact
          value: /ready
    backendRefs:
      - name: health-service
        port: 8080
```

## Weighted Backends

Split traffic between multiple backends:

```yaml
rules:
  - backendRefs:
      - name: service-v1
        port: 8080
        weight: 90
      - name: service-v2
        port: 8080
        weight: 10
```

## Request Timeouts

Configure per-rule request timeouts:

```yaml
rules:
  - matches:
      - path:
          type: PathPrefix
          value: /slow-api
    backendRefs:
      - name: slow-service
        port: 8080
    timeouts:
      request: "60s"
```

## Multiple Hostnames

Route multiple hostnames to the same backend:

```yaml
spec:
  hostnames:
    - app.example.com
    - www.example.com
    - "*.staging.example.com"
  rules:
    - backendRefs:
        - name: web-app
          port: 80
```

## Cross-Namespace Backend

Reference a service in another namespace (requires ReferenceGrant):

```yaml
rules:
  - backendRefs:
      - name: shared-service
        namespace: shared-services
        port: 8080
```

See [ReferenceGrant](referencegrant.md) for setup.

## Complete Example

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: comprehensive-route
  namespace: default
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - api.example.com
  rules:
    # Health check endpoint
    - matches:
        - path:
            type: Exact
            value: /health
      backendRefs:
        - name: health-service
          port: 8080
      timeouts:
        request: "5s"

    # API v2 with canary
    - matches:
        - path:
            type: PathPrefix
            value: /api/v2
      backendRefs:
        - name: api-v2-stable
          port: 8080
          weight: 95
        - name: api-v2-canary
          port: 8080
          weight: 5

    # API v1 (default)
    - matches:
        - path:
            type: PathPrefix
            value: /api
      backendRefs:
        - name: api-v1
          port: 8080
      timeouts:
        request: "30s"

    # Default fallback
    - backendRefs:
        - name: default-backend
          port: 80
```

## Next Steps

- Configure [GRPCRoute](grpcroute.md) for gRPC services
- Set up [ReferenceGrant](referencegrant.md) for cross-namespace routing
- Review [Limitations](limitations.md) for unsupported features
