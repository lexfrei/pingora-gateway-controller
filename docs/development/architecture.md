# Architecture

Deep dive into Pingora Gateway Controller architecture and design.

## High-Level Architecture

```mermaid
graph TB
    subgraph Kubernetes Cluster
        subgraph Control Plane
            API[Kubernetes API Server]
        end

        subgraph pingora-system
            CTL[Controller<br/>Go]
            PX[Pingora Proxy<br/>Rust]
        end

        subgraph Application Namespaces
            GW[Gateway]
            HR[HTTPRoute]
            GR[GRPCRoute]
            SVC[Backend Services]
        end
    end

    CLIENT[Client Traffic] -->|HTTP/gRPC| PX
    API -->|watch| CTL
    CTL -->|gRPC sync| PX
    PX -->|proxy| SVC
    HR -->|parentRef| GW
    GR -->|parentRef| GW
```

## Components

### Controller (Go)

The controller is a Kubernetes operator built with controller-runtime:

```mermaid
graph LR
    subgraph Controller
        GWC[GatewayReconciler]
        HRC[HTTPRouteReconciler]
        GRC[GRPCRouteReconciler]
        SYNC[PingoraSyncer]
        BUILD[PingoraBuilder]
    end

    API[K8s API] --> GWC
    API --> HRC
    API --> GRC
    HRC --> BUILD
    GRC --> BUILD
    BUILD --> SYNC
    SYNC -->|gRPC| PROXY[Pingora Proxy]
```

### Pingora Proxy (Rust)

The proxy is built on Cloudflare's Pingora framework:

- High-performance HTTP/gRPC reverse proxy
- Dynamic configuration via gRPC API
- Zero-downtime route updates

## Controller Components

### GatewayReconciler

Watches Gateway resources and manages their lifecycle:

- Validates GatewayClass reference
- Resolves PingoraConfig from parametersRef
- Updates Gateway status conditions

### HTTPRouteReconciler

Watches HTTPRoute resources:

- Validates parent Gateway references
- Resolves backend Service references
- Triggers route synchronization
- Updates HTTPRoute status

### GRPCRouteReconciler

Similar to HTTPRouteReconciler for GRPCRoute:

- Validates parent Gateway references
- Resolves backend Service references
- Triggers route synchronization
- Updates GRPCRoute status

### PingoraSyncer

Manages communication with Pingora proxy:

- Establishes gRPC connection
- Converts routes to protobuf format
- Sends configuration updates
- Handles connection retry logic

### PingoraBuilder

Converts Gateway API resources to Pingora format:

- Builds route match conditions
- Resolves backend addresses
- Applies timeout configuration

## Data Flow

### Route Configuration Flow

```mermaid
sequenceDiagram
    participant User
    participant K8s as Kubernetes API
    participant Ctrl as Controller
    participant Builder as PingoraBuilder
    participant Syncer as PingoraSyncer
    participant Proxy as Pingora Proxy

    User->>K8s: Create HTTPRoute
    K8s->>Ctrl: Watch event
    Ctrl->>K8s: Get referenced Gateway
    Ctrl->>K8s: Get backend Services
    Ctrl->>Builder: Build Pingora route
    Builder-->>Ctrl: Route config
    Ctrl->>Syncer: Sync routes
    Syncer->>Proxy: gRPC SyncRoutes
    Proxy-->>Syncer: Success
    Syncer-->>Ctrl: Sync complete
    Ctrl->>K8s: Update HTTPRoute status
```

### Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Proxy as Pingora Proxy
    participant Backend as Backend Service

    Client->>Proxy: HTTP Request
    Note over Proxy: Match route by<br/>host, path, headers
    Proxy->>Backend: Forward request
    Backend-->>Proxy: Response
    Proxy-->>Client: Response
```

## Resource Relationships

```mermaid
erDiagram
    GatewayClass ||--o| PingoraConfig : "parametersRef"
    Gateway }|--|| GatewayClass : "gatewayClassName"
    HTTPRoute }o--|| Gateway : "parentRef"
    GRPCRoute }o--|| Gateway : "parentRef"
    HTTPRoute }o--|{ Service : "backendRef"
    GRPCRoute }o--|{ Service : "backendRef"
    ReferenceGrant ||--o{ HTTPRoute : "allows"
    ReferenceGrant ||--o{ GRPCRoute : "allows"
```

## Key Design Decisions

### Why gRPC for Controller-Proxy Communication?

- Efficient binary protocol
- Strong typing with protobuf
- Bi-directional streaming capability
- Built-in health checking

### Why Separate Controller and Proxy?

- Independent scaling
- Language-specific optimization (Go for K8s, Rust for performance)
- Independent deployment lifecycle
- Clear separation of concerns

### Why controller-runtime?

- Battle-tested Kubernetes controller framework
- Built-in leader election
- Efficient watch caching
- Standardized patterns

## Package Structure

```text
internal/
├── config/
│   └── pingora_resolver.go    # PingoraConfig resolution
├── controller/
│   ├── gateway_controller.go  # Gateway reconciler
│   ├── httproute_controller.go # HTTPRoute reconciler
│   ├── grpcroute_controller.go # GRPCRoute reconciler
│   └── pingora_syncer.go      # gRPC sync logic
├── dns/
│   └── detect.go              # Cluster domain detection
├── ingress/
│   └── pingora_builder.go     # Route conversion
└── metrics/
    └── metrics.go             # Prometheus metrics
```

## Configuration Flow

```mermaid
graph TD
    GC[GatewayClass] -->|parametersRef| PC[PingoraConfig]
    PC -->|address| GRPC[gRPC Connection]
    PC -->|tls| TLS[TLS Config]
    PC -->|connection| CONN[Connection Params]
    GRPC --> PROXY[Pingora Proxy]
```

## Error Handling

### Retry Strategy

- gRPC connection: Exponential backoff with jitter
- Failed syncs: Immediate retry with rate limiting
- Transient errors: Automatic retry via controller-runtime

### Status Reporting

- Gateway conditions: Accepted, Programmed
- Route conditions: Accepted, ResolvedRefs
- PingoraConfig status: Connected, LastSyncTime

## Next Steps

- Read [Contributing Guidelines](contributing.md)
- Learn about [Testing](testing.md)
