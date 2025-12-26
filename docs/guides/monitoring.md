# Monitoring

This guide covers setting up Prometheus metrics collection and Grafana
dashboards for Pingora Gateway Controller.

## Overview

The controller exposes Prometheus metrics for:

- Route synchronization performance
- gRPC communication with Pingora proxy
- Backend reference validation
- Error tracking

## Prometheus Setup

### Enable ServiceMonitor

If using Prometheus Operator, enable ServiceMonitor in Helm values:

```yaml
serviceMonitor:
  enabled: true
  interval: "30s"
  labels:
    release: prometheus  # Match your Prometheus selector
```

### Manual Prometheus Configuration

For standard Prometheus installations, add a scrape config:

```yaml
scrape_configs:
  - job_name: 'pingora-gateway-controller'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
            - pingora-system
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: pingora-gateway-controller
      - source_labels: [__meta_kubernetes_endpoint_port_name]
        action: keep
        regex: metrics
```

### Verify Metrics

Check metrics are being scraped:

```bash
kubectl port-forward --namespace pingora-system \
  service/pingora-gateway-controller 8080:8080

curl http://localhost:8080/metrics
```

## Available Metrics

### Sync Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `pingora_sync_duration_seconds` | Histogram | Duration of route synchronization |
| `pingora_synced_routes` | Gauge | Number of synced routes by type |
| `pingora_ingress_rules` | Gauge | Total ingress rules in proxy config |
| `pingora_failed_backend_refs` | Gauge | Failed backend references by type |
| `pingora_sync_errors_total` | Counter | Total sync errors by type |

### Ingress Build Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `pingora_ingress_build_duration_seconds` | Histogram | Duration of ingress rule building |
| `pingora_backend_ref_validation_total` | Counter | Backend ref validation results |

### gRPC Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `pingora_grpc_duration_seconds` | Histogram | Duration of gRPC calls |
| `pingora_grpc_calls_total` | Counter | Total gRPC calls by method and status |
| `pingora_grpc_errors_total` | Counter | Total gRPC errors by method and type |

## Alerting Rules

### Example PrometheusRule

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: pingora-gateway-alerts
  namespace: pingora-system
  labels:
    release: prometheus
spec:
  groups:
    - name: pingora-gateway
      rules:
        # Sync taking too long
        - alert: PingoraSyncSlow
          expr: |
            histogram_quantile(0.95,
              sum(rate(pingora_sync_duration_seconds_bucket[5m])) by (le)
            ) > 5
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Pingora sync is slow"
            description: "95th percentile sync duration is {{ $value }}s"

        # High error rate
        - alert: PingoraSyncErrors
          expr: |
            increase(pingora_sync_errors_total[5m]) > 10
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "High Pingora sync error rate"
            description: "{{ $value }} sync errors in last 5 minutes"

        # gRPC connection issues
        - alert: PingoraGRPCErrors
          expr: |
            increase(pingora_grpc_errors_total[5m]) > 5
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Pingora gRPC errors detected"
            description: "{{ $value }} gRPC errors in last 5 minutes"

        # No routes synced
        - alert: PingoraNoRoutes
          expr: |
            sum(pingora_synced_routes) == 0
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "No routes synced"
            description: "Controller has no synced routes for 10 minutes"

        # Failed backend references
        - alert: PingoraFailedBackends
          expr: |
            sum(pingora_failed_backend_refs) > 0
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Failed backend references"
            description: "{{ $value }} backend references failed"
```

## Grafana Dashboard

### Import Dashboard

Create a dashboard with the following panels:

#### Sync Performance

```json
{
  "title": "Sync Duration (p95)",
  "type": "graph",
  "targets": [
    {
      "expr": "histogram_quantile(0.95, sum(rate(pingora_sync_duration_seconds_bucket[5m])) by (le))",
      "legendFormat": "p95"
    },
    {
      "expr": "histogram_quantile(0.50, sum(rate(pingora_sync_duration_seconds_bucket[5m])) by (le))",
      "legendFormat": "p50"
    }
  ]
}
```

#### Routes Overview

```json
{
  "title": "Synced Routes",
  "type": "stat",
  "targets": [
    {
      "expr": "sum(pingora_synced_routes) by (type)",
      "legendFormat": "{{ type }}"
    }
  ]
}
```

#### gRPC Calls Rate

```json
{
  "title": "gRPC Calls/sec",
  "type": "graph",
  "targets": [
    {
      "expr": "sum(rate(pingora_grpc_calls_total[5m])) by (method, status)",
      "legendFormat": "{{ method }} ({{ status }})"
    }
  ]
}
```

#### Error Rate

```json
{
  "title": "Errors",
  "type": "graph",
  "targets": [
    {
      "expr": "sum(rate(pingora_sync_errors_total[5m])) by (error_type)",
      "legendFormat": "Sync: {{ error_type }}"
    },
    {
      "expr": "sum(rate(pingora_grpc_errors_total[5m])) by (error_type)",
      "legendFormat": "gRPC: {{ error_type }}"
    }
  ]
}
```

## Useful Queries

### Sync Success Rate

```promql
sum(rate(pingora_grpc_calls_total{status="success"}[5m])) /
sum(rate(pingora_grpc_calls_total[5m]))
```

### Average Sync Duration

```promql
sum(rate(pingora_sync_duration_seconds_sum[5m])) /
sum(rate(pingora_sync_duration_seconds_count[5m]))
```

### Routes by Type

```promql
sum(pingora_synced_routes) by (type)
```

### Backend Validation Failures

```promql
sum(rate(pingora_backend_ref_validation_total{result="failure"}[5m])) by (reason)
```

## Proxy Metrics

The Pingora proxy exposes its own metrics. See proxy documentation for
available metrics and scrape configuration.

## Next Steps

- Set up [Cross-Namespace Routing](cross-namespace.md) for multi-tenant environments
- Check [Operations](../operations/index.md) for troubleshooting
- Review [Metrics Reference](../operations/metrics.md) for complete metric list
