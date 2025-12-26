# Controller Options

The controller accepts configuration via CLI flags and environment variables.

## CLI Flags

### Core Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gateway-class-name` | `pingora` | GatewayClass name to watch |
| `--controller-name` | `pingora.k8s.lex.la/gateway-controller` | Controller identifier for GatewayClass |
| `--cluster-domain` | auto-detected | Kubernetes cluster domain for DNS |

### Observability Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--metrics-addr` | `:8080` | Address for Prometheus metrics endpoint |
| `--health-addr` | `:8081` | Address for health probe endpoints |
| `--log-level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `--log-format` | `json` | Log format: `json`, `text` |

### Leader Election Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--leader-elect` | `false` | Enable leader election for HA |
| `--leader-election-namespace` | controller namespace | Namespace for leader election lease |
| `--leader-election-name` | `pingora-gateway-controller-leader` | Name of the lease resource |

## Environment Variables

All flags can be set via environment variables with `PINGORA_` prefix:

| Environment Variable | CLI Flag |
|---------------------|----------|
| `PINGORA_GATEWAY_CLASS_NAME` | `--gateway-class-name` |
| `PINGORA_CONTROLLER_NAME` | `--controller-name` |
| `PINGORA_CLUSTER_DOMAIN` | `--cluster-domain` |
| `PINGORA_METRICS_ADDR` | `--metrics-addr` |
| `PINGORA_HEALTH_ADDR` | `--health-addr` |
| `PINGORA_LOG_LEVEL` | `--log-level` |
| `PINGORA_LOG_FORMAT` | `--log-format` |
| `PINGORA_LEADER_ELECT` | `--leader-elect` |

!!! note "Precedence"

    CLI flags take precedence over environment variables.

## Cluster Domain Detection

The controller automatically detects the Kubernetes cluster domain from
`/etc/resolv.conf` search domains. Override with `--cluster-domain` if:

- Auto-detection fails
- Using a non-standard cluster domain
- Running outside the cluster

```bash
# Check auto-detected domain
kubectl logs deployment/pingora-gateway-controller | grep "cluster domain"
```

## Health Endpoints

The controller exposes health endpoints on `--health-addr`:

| Endpoint | Description |
|----------|-------------|
| `/healthz` | Liveness probe - controller is running |
| `/readyz` | Readiness probe - controller can accept traffic |

## Metrics Endpoint

Prometheus metrics are exposed on `--metrics-addr`:

```bash
curl http://localhost:8080/metrics
```

See [Metrics Reference](../operations/metrics.md) for available metrics.

## Example: Full Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pingora-gateway-controller
spec:
  template:
    spec:
      containers:
        - name: controller
          args:
            - --gateway-class-name=pingora
            - --controller-name=pingora.k8s.lex.la/gateway-controller
            - --metrics-addr=:8080
            - --health-addr=:8081
            - --log-level=debug
            - --log-format=json
            - --leader-elect=true
            - --leader-election-namespace=pingora-system
          env:
            # Environment variables can also be used
            - name: PINGORA_LOG_LEVEL
              value: "debug"
```

## Next Steps

- Configure [Helm Values](helm-values.md) for deployment customization
- Set up [PingoraConfig](gatewayclassconfig.md) for proxy connection
