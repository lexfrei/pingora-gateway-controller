# Helm Chart Reference

Complete reference for the Pingora Gateway Controller Helm chart.

## Installation

```bash
helm install pingora-gateway-controller \
  oci://ghcr.io/lexfrei/pingora-gateway-controller/chart \
  --namespace pingora-system \
  --create-namespace
```

## Chart Info

| Property | Value |
|----------|-------|
| Chart Name | `pingora-gateway-controller` |
| Chart Type | application |
| Home | https://github.com/lexfrei/pingora-gateway-controller |

## Values Reference

### Global

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `nameOverride` | string | `""` | Override chart name |
| `fullnameOverride` | string | `""` | Override full release name |
| `imagePullSecrets` | list | `[]` | Image pull secrets |

### Controller Image

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `image.repository` | string | `ghcr.io/lexfrei/pingora-gateway-controller` | Image repository |
| `image.pullPolicy` | string | `IfNotPresent` | Image pull policy |
| `image.tag` | string | `""` | Image tag (defaults to appVersion) |

### Controller Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `controller.gatewayClassName` | string | `pingora` | GatewayClass name to watch |
| `controller.controllerName` | string | `pingora.k8s.lex.la/gateway-controller` | Controller identifier |
| `controller.clusterDomain` | string | `""` | Cluster domain (auto-detected) |
| `controller.logLevel` | string | `info` | Log level: debug, info, warn, error |
| `controller.logFormat` | string | `json` | Log format: json, text |

### Leader Election

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `leaderElection.enabled` | bool | `false` | Enable leader election |
| `leaderElection.namespace` | string | `""` | Lease namespace |
| `leaderElection.leaseName` | string | `pingora-gateway-controller-leader` | Lease name |

### PingoraConfig

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `pingoraConfig.create` | bool | `true` | Create PingoraConfig resource |
| `pingoraConfig.name` | string | `""` | Config name (defaults to release name) |
| `pingoraConfig.address` | string | `""` | Proxy gRPC address (auto-configured) |

### PingoraConfig TLS

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `pingoraConfig.tls.enabled` | bool | `false` | Enable TLS |
| `pingoraConfig.tls.secretRef.name` | string | `""` | TLS Secret name |
| `pingoraConfig.tls.secretRef.namespace` | string | `""` | TLS Secret namespace |
| `pingoraConfig.tls.insecureSkipVerify` | bool | `false` | Skip TLS verification |
| `pingoraConfig.tls.serverName` | string | `""` | TLS server name override |

### PingoraConfig Connection

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `pingoraConfig.connection.connectTimeoutSeconds` | int | `5` | Connection timeout |
| `pingoraConfig.connection.requestTimeoutSeconds` | int | `30` | Request timeout |
| `pingoraConfig.connection.keepaliveTimeSeconds` | int | `30` | Keepalive interval |
| `pingoraConfig.connection.maxRetries` | int | `3` | Max retry attempts |
| `pingoraConfig.connection.retryBackoffMs` | int | `1000` | Retry backoff (ms) |

### Proxy

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `proxy.enabled` | bool | `true` | Deploy proxy |
| `proxy.replicaCount` | int | `2` | Proxy replicas |
| `proxy.image.repository` | string | `ghcr.io/lexfrei/pingora-proxy` | Proxy image |
| `proxy.image.pullPolicy` | string | `IfNotPresent` | Pull policy |
| `proxy.image.tag` | string | `""` | Image tag |
| `proxy.logLevel` | string | `info` | Log level |
| `proxy.service.type` | string | `ClusterIP` | Service type |
| `proxy.service.annotations` | object | `{}` | Service annotations |

### Proxy Resources

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `proxy.resources.limits.cpu` | string | `500m` | CPU limit |
| `proxy.resources.limits.memory` | string | `512Mi` | Memory limit |
| `proxy.resources.requests.cpu` | string | `100m` | CPU request |
| `proxy.resources.requests.memory` | string | `128Mi` | Memory request |

### Replica Count

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `replicaCount` | int | `1` | Controller replicas |

### Resources

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `resources.limits.cpu` | string | `200m` | CPU limit |
| `resources.limits.memory` | string | `256Mi` | Memory limit |
| `resources.requests.cpu` | string | `100m` | CPU request |
| `resources.requests.memory` | string | `128Mi` | Memory request |

### Security Context

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `podSecurityContext.runAsNonRoot` | bool | `true` | Run as non-root |
| `podSecurityContext.runAsUser` | int | `65534` | User ID |
| `podSecurityContext.seccompProfile.type` | string | `RuntimeDefault` | Seccomp profile |
| `securityContext.allowPrivilegeEscalation` | bool | `false` | Disable privilege escalation |
| `securityContext.readOnlyRootFilesystem` | bool | `true` | Read-only filesystem |
| `securityContext.capabilities.drop` | list | `[ALL]` | Drop all capabilities |

### Service Account

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `serviceAccount.annotations` | object | `{}` | SA annotations |
| `serviceAccount.name` | string | `""` | SA name |

### Service

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `service.type` | string | `ClusterIP` | Service type |
| `service.metricsPort` | int | `8080` | Metrics port |
| `service.healthPort` | int | `8081` | Health port |
| `service.annotations` | object | `{}` | Annotations |

### ServiceMonitor

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `serviceMonitor.enabled` | bool | `false` | Create ServiceMonitor |
| `serviceMonitor.interval` | string | `""` | Scrape interval |
| `serviceMonitor.labels` | object | `{}` | Additional labels |

### Health Probes

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `healthProbes.startupProbe.enabled` | bool | `true` | Enable startup probe |
| `healthProbes.startupProbe.initialDelaySeconds` | int | `0` | Initial delay |
| `healthProbes.startupProbe.periodSeconds` | int | `5` | Period |
| `healthProbes.startupProbe.failureThreshold` | int | `12` | Failure threshold |
| `healthProbes.livenessProbe.enabled` | bool | `true` | Enable liveness |
| `healthProbes.livenessProbe.initialDelaySeconds` | int | `15` | Initial delay |
| `healthProbes.livenessProbe.periodSeconds` | int | `20` | Period |
| `healthProbes.readinessProbe.enabled` | bool | `true` | Enable readiness |
| `healthProbes.readinessProbe.initialDelaySeconds` | int | `5` | Initial delay |
| `healthProbes.readinessProbe.periodSeconds` | int | `10` | Period |

### Pod Disruption Budget

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `podDisruptionBudget.enabled` | bool | `false` | Enable PDB |
| `podDisruptionBudget.minAvailable` | int | `1` | Min available |
| `podDisruptionBudget.maxUnavailable` | - | `null` | Max unavailable |

### Network Policy

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `networkPolicy.enabled` | bool | `false` | Enable NetworkPolicy |
| `networkPolicy.ingress.from` | list | `[]` | Ingress sources |
| `networkPolicy.pingoraProxy.port` | int | `50051` | Proxy gRPC port |

### GatewayClass

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `gatewayClass.create` | bool | `true` | Create GatewayClass |

### Scheduling

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `nodeSelector` | object | `{}` | Node selector |
| `tolerations` | list | `[]` | Tolerations |
| `affinity` | object | `{}` | Affinity rules |
| `topologySpreadConstraints` | list | `[]` | Topology spread |
| `priorityClassName` | string | `""` | Priority class |

## Example Values Files

### Minimal

```yaml
# minimal-values.yaml
controller:
  logLevel: info
```

### Production

```yaml
# production-values.yaml
replicaCount: 2

leaderElection:
  enabled: true

controller:
  logLevel: info
  logFormat: json

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

### With TLS

```yaml
# tls-values.yaml
pingoraConfig:
  tls:
    enabled: true
    secretRef:
      name: pingora-tls
      namespace: pingora-system
    serverName: pingora-proxy
```

## Next Steps

- Review [CRD Reference](crd-reference.md)
- Check [Security](security.md) best practices
