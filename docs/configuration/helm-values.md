# Helm Values

Complete reference for Helm chart configuration values.

## Controller Configuration

### `controller`

Controller behavior settings.

```yaml
controller:
  # GatewayClass name to watch
  gatewayClassName: "pingora"

  # Controller identifier (must be unique in cluster)
  controllerName: "pingora.k8s.lex.la/gateway-controller"

  # Kubernetes cluster domain (auto-detected if empty)
  clusterDomain: ""

  # Log level: debug, info, warn, error
  logLevel: "info"

  # Log format: json, text
  logFormat: "json"
```

### `leaderElection`

High availability settings for multi-replica deployments.

```yaml
leaderElection:
  # Enable leader election
  enabled: false

  # Namespace for leader election lease
  namespace: ""

  # Name of the lease resource
  leaseName: "pingora-gateway-controller-leader"
```

!!! warning "Required for HA"

    Leader election must be enabled when running multiple controller replicas.

## PingoraConfig Settings

### `pingoraConfig`

Configuration for the PingoraConfig CRD resource.

```yaml
pingoraConfig:
  # Create PingoraConfig resource
  create: true

  # Name of the PingoraConfig (defaults to release name)
  name: ""

  # gRPC endpoint address of Pingora proxy
  # Auto-configured when proxy.enabled=true and address is empty
  address: ""

  # TLS settings for gRPC connection
  tls:
    enabled: false
    secretRef:
      name: ""
      namespace: ""
    insecureSkipVerify: false
    serverName: ""

  # Connection parameters
  connection:
    connectTimeoutSeconds: 5
    requestTimeoutSeconds: 30
    keepaliveTimeSeconds: 30
    maxRetries: 3
    retryBackoffMs: 1000
```

## Proxy Configuration

### `proxy`

Pingora proxy deployment settings.

```yaml
proxy:
  # Enable proxy deployment
  enabled: true

  # Number of proxy replicas
  replicaCount: 2

  # Container image
  image:
    repository: ghcr.io/lexfrei/pingora-proxy
    pullPolicy: IfNotPresent
    tag: ""  # Defaults to appVersion

  # Log level: trace, debug, info, warn, error
  logLevel: "info"

  # Resource limits
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi

  # Service configuration
  service:
    type: ClusterIP
    annotations: {}
```

## Image Configuration

### `image`

Controller container image settings.

```yaml
image:
  repository: ghcr.io/lexfrei/pingora-gateway-controller
  pullPolicy: IfNotPresent
  tag: ""  # Defaults to appVersion

imagePullSecrets: []
```

## Resource Management

### `resources`

Controller resource requests and limits.

```yaml
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### `replicaCount`

Number of controller replicas. Enable `leaderElection` for multiple replicas.

```yaml
replicaCount: 1
```

## Security Settings

### `podSecurityContext`

Pod-level security settings.

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65534
  seccompProfile:
    type: RuntimeDefault
```

### `securityContext`

Container-level security settings.

```yaml
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
```

## Service and Networking

### `service`

Controller service configuration.

```yaml
service:
  type: ClusterIP
  metricsPort: 8080
  healthPort: 8081
  annotations: {}
```

### `networkPolicy`

Network policy for controller pods.

```yaml
networkPolicy:
  enabled: false
  ingress:
    from: []
  pingoraProxy:
    port: 50051
    namespaceSelector: {}
    podSelector: {}
```

## Observability

### `serviceMonitor`

Prometheus ServiceMonitor configuration.

```yaml
serviceMonitor:
  enabled: false
  interval: ""  # Uses Prometheus default
  labels: {}    # For Prometheus selector
```

## Health Probes

### `healthProbes`

Kubernetes health probe settings.

```yaml
healthProbes:
  startupProbe:
    enabled: true
    initialDelaySeconds: 0
    periodSeconds: 5
    timeoutSeconds: 3
    failureThreshold: 12

  livenessProbe:
    enabled: true
    initialDelaySeconds: 15
    periodSeconds: 20
    timeoutSeconds: 5
    failureThreshold: 3

  readinessProbe:
    enabled: true
    initialDelaySeconds: 5
    periodSeconds: 10
    timeoutSeconds: 3
    failureThreshold: 3
```

## High Availability

### `podDisruptionBudget`

PDB settings for availability during disruptions.

```yaml
podDisruptionBudget:
  enabled: false
  minAvailable: 1
  maxUnavailable: null
  unhealthyPodEvictionPolicy: "IfHealthyBudget"
```

## Scheduling

### Node Selection and Affinity

```yaml
nodeSelector: {}
tolerations: []
affinity: {}
topologySpreadConstraints: []
priorityClassName: ""
```

## GatewayClass

### `gatewayClass`

GatewayClass resource creation.

```yaml
gatewayClass:
  # Create GatewayClass resource
  create: true
```

## Example: Production Configuration

```yaml
# Production-ready values
replicaCount: 2

leaderElection:
  enabled: true

controller:
  logLevel: "info"
  logFormat: "json"

proxy:
  enabled: true
  replicaCount: 3
  resources:
    limits:
      cpu: 2
      memory: 2Gi
    requests:
      cpu: 500m
      memory: 512Mi

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 200m
    memory: 256Mi

podDisruptionBudget:
  enabled: true
  minAvailable: 1

serviceMonitor:
  enabled: true
  labels:
    release: prometheus

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app.kubernetes.io/name: pingora-gateway-controller
          topologyKey: kubernetes.io/hostname
```

## Next Steps

- Configure [PingoraConfig](gatewayclassconfig.md) for proxy connection
- Learn about [Gateway API](../gateway-api/index.md) resources
