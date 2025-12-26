# Operations

This section covers operational aspects of running Pingora Gateway Controller
in production.

## Sections

<div class="grid cards" markdown>

-   :material-wrench:{ .lg .middle } **Troubleshooting**

    ---

    Common issues and their solutions for debugging controller and proxy problems.

    [:octicons-arrow-right-24: Troubleshooting](troubleshooting.md)

-   :material-chart-box:{ .lg .middle } **Metrics Reference**

    ---

    Complete reference of all Prometheus metrics exposed by the controller.

    [:octicons-arrow-right-24: Metrics](metrics.md)

-   :material-file-document:{ .lg .middle } **Manual Installation**

    ---

    Step-by-step guide for installing without Helm using raw Kubernetes manifests.

    [:octicons-arrow-right-24: Manual Installation](manual-installation.md)

</div>

## Quick Commands

### Check Controller Status

```bash
kubectl get pods --namespace pingora-system
kubectl logs --selector app.kubernetes.io/name=pingora-gateway-controller \
  --namespace pingora-system --tail=100
```

### Check Route Status

```bash
kubectl get httproutes --all-namespaces
kubectl get grpcroutes --all-namespaces
```

### Check Gateway Status

```bash
kubectl get gateway --namespace pingora-system --output yaml
```

### Check PingoraConfig

```bash
kubectl get pingoraconfig --output wide
```

## Health Checks

### Controller Health

```bash
kubectl exec -it deployment/pingora-gateway-controller \
  --namespace pingora-system -- wget -qO- http://localhost:8081/healthz
```

### Proxy Health

```bash
kubectl exec -it deployment/pingora-gateway-controller-proxy \
  --namespace pingora-system -- wget -qO- http://localhost:8081/health
```

## Next Steps

- Review [Troubleshooting](troubleshooting.md) for common issues
- Check [Metrics](metrics.md) for monitoring setup
