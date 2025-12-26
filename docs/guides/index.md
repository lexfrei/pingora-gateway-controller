# Guides

Practical guides for common tasks and production deployments.

## Available Guides

<div class="grid cards" markdown>

-   :material-share-variant:{ .lg .middle } **Cross-Namespace Routing**

    ---

    Configure ReferenceGrant for cross-namespace service references
    and multi-tenant deployments.

    [:octicons-arrow-right-24: Cross-Namespace](cross-namespace.md)

-   :material-chart-line:{ .lg .middle } **Monitoring**

    ---

    Set up Prometheus metrics collection and Grafana dashboards
    for production visibility.

    [:octicons-arrow-right-24: Monitoring](monitoring.md)

</div>

## Quick Tips

### High Availability

For production deployments:

```yaml
replicaCount: 2
leaderElection:
  enabled: true
proxy:
  replicaCount: 3
podDisruptionBudget:
  enabled: true
  minAvailable: 1
```

### Resource Sizing

Start with default resources and adjust based on:

- Number of routes and backends
- Request throughput
- Metrics from monitoring

### Security

- Enable NetworkPolicy to restrict traffic
- Use ReferenceGrant for cross-namespace access
- Run as non-root (default configuration)

## Next Steps

- Set up [Cross-Namespace Routing](cross-namespace.md) for multi-tenant environments
- Configure [Monitoring](monitoring.md) for production observability
