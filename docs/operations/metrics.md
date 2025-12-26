# Metrics Reference

Complete reference of Prometheus metrics exposed by Pingora Gateway Controller.

## Endpoints

| Component | Port | Path |
|-----------|------|------|
| Controller | 8080 | `/metrics` |

## Sync Metrics

### pingora_sync_duration_seconds

Duration of route synchronization to Pingora proxy.

| Label | Description |
|-------|-------------|
| `status` | Sync result: `success`, `error` |

**Type**: Histogram

**Buckets**: 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30 seconds

**Example**:

```promql
# 95th percentile sync duration
histogram_quantile(0.95,
  sum(rate(pingora_sync_duration_seconds_bucket[5m])) by (le)
)

# Average sync duration
sum(rate(pingora_sync_duration_seconds_sum[5m])) /
sum(rate(pingora_sync_duration_seconds_count[5m]))
```

### pingora_synced_routes

Number of routes synced by type.

| Label | Description |
|-------|-------------|
| `type` | Route type: `HTTPRoute`, `GRPCRoute` |

**Type**: Gauge

**Example**:

```promql
# Total routes by type
sum(pingora_synced_routes) by (type)

# Total routes across all types
sum(pingora_synced_routes)
```

### pingora_ingress_rules

Total ingress rules in proxy configuration.

**Type**: Gauge

**Example**:

```promql
pingora_ingress_rules
```

### pingora_failed_backend_refs

Number of failed backend references by route type.

| Label | Description |
|-------|-------------|
| `type` | Route type: `HTTPRoute`, `GRPCRoute` |

**Type**: Gauge

**Example**:

```promql
# Alert on failed backends
sum(pingora_failed_backend_refs) > 0
```

### pingora_sync_errors_total

Total sync errors by type.

| Label | Description |
|-------|-------------|
| `error_type` | Error category |

**Type**: Counter

**Example**:

```promql
# Error rate per minute
sum(rate(pingora_sync_errors_total[1m])) by (error_type)
```

## Ingress Build Metrics

### pingora_ingress_build_duration_seconds

Duration of ingress rule building.

| Label | Description |
|-------|-------------|
| `type` | Route type: `HTTPRoute`, `GRPCRoute` |

**Type**: Histogram

**Buckets**: 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5 seconds

**Example**:

```promql
# Average build time by type
sum(rate(pingora_ingress_build_duration_seconds_sum[5m])) by (type) /
sum(rate(pingora_ingress_build_duration_seconds_count[5m])) by (type)
```

### pingora_backend_ref_validation_total

Backend reference validation results.

| Label | Description |
|-------|-------------|
| `type` | Route type: `HTTPRoute`, `GRPCRoute` |
| `result` | Validation result: `success`, `failure` |
| `reason` | Failure reason (if applicable) |

**Type**: Counter

**Example**:

```promql
# Validation failure rate
sum(rate(pingora_backend_ref_validation_total{result="failure"}[5m])) by (reason)

# Success rate
sum(rate(pingora_backend_ref_validation_total{result="success"}[5m])) /
sum(rate(pingora_backend_ref_validation_total[5m]))
```

## gRPC Metrics

### pingora_grpc_duration_seconds

Duration of gRPC calls to Pingora proxy.

| Label | Description |
|-------|-------------|
| `method` | gRPC method name |

**Type**: Histogram

**Buckets**: 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5 seconds

**Example**:

```promql
# 99th percentile gRPC latency
histogram_quantile(0.99,
  sum(rate(pingora_grpc_duration_seconds_bucket[5m])) by (le, method)
)
```

### pingora_grpc_calls_total

Total gRPC calls to Pingora proxy.

| Label | Description |
|-------|-------------|
| `method` | gRPC method name |
| `status` | Call status: `success`, `error` |

**Type**: Counter

**Example**:

```promql
# Call rate by method and status
sum(rate(pingora_grpc_calls_total[5m])) by (method, status)

# Success rate
sum(rate(pingora_grpc_calls_total{status="success"}[5m])) /
sum(rate(pingora_grpc_calls_total[5m]))
```

### pingora_grpc_errors_total

Total gRPC errors by type.

| Label | Description |
|-------|-------------|
| `method` | gRPC method name |
| `error_type` | Error category |

**Type**: Counter

**Example**:

```promql
# Error rate by method
sum(rate(pingora_grpc_errors_total[5m])) by (method, error_type)
```

## Controller-Runtime Metrics

The controller also exposes standard controller-runtime metrics:

### controller_runtime_reconcile_total

Total reconciliation attempts.

| Label | Description |
|-------|-------------|
| `controller` | Controller name |
| `result` | `success`, `error`, `requeue` |

### controller_runtime_reconcile_time_seconds

Time taken by reconciliation.

| Label | Description |
|-------|-------------|
| `controller` | Controller name |

### workqueue_depth

Current depth of the work queue.

### workqueue_adds_total

Total elements added to the queue.

## Recommended Alerts

```yaml
groups:
  - name: pingora-gateway
    rules:
      - alert: PingoraSyncErrors
        expr: increase(pingora_sync_errors_total[5m]) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Sync errors detected"

      - alert: PingoraGRPCErrors
        expr: increase(pingora_grpc_errors_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "gRPC communication errors"

      - alert: PingoraSyncSlow
        expr: |
          histogram_quantile(0.95,
            sum(rate(pingora_sync_duration_seconds_bucket[5m])) by (le)
          ) > 10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Slow route synchronization"

      - alert: PingoraFailedBackends
        expr: sum(pingora_failed_backend_refs) > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Failed backend references detected"
```

## Grafana Queries

### Key Dashboard Panels

```promql
# Routes Overview
sum(pingora_synced_routes) by (type)

# Sync Performance
histogram_quantile(0.95, sum(rate(pingora_sync_duration_seconds_bucket[5m])) by (le))

# gRPC Success Rate
sum(rate(pingora_grpc_calls_total{status="success"}[5m])) /
sum(rate(pingora_grpc_calls_total[5m])) * 100

# Error Rate
sum(rate(pingora_sync_errors_total[5m])) by (error_type)
```

## Next Steps

- Set up [Monitoring](../guides/monitoring.md) with Prometheus and Grafana
- Review [Troubleshooting](troubleshooting.md) for common issues
