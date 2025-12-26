# Troubleshooting

Common issues and solutions for Pingora Gateway Controller.

## Diagnostic Commands

### Check Component Status

```bash
# Controller pods
kubectl get pods --namespace pingora-system \
  --selector app.kubernetes.io/name=pingora-gateway-controller

# Proxy pods
kubectl get pods --namespace pingora-system \
  --selector app.kubernetes.io/component=proxy

# All resources
kubectl get all --namespace pingora-system
```

### View Logs

```bash
# Controller logs
kubectl logs --namespace pingora-system \
  --selector app.kubernetes.io/name=pingora-gateway-controller \
  --tail=100 --follow

# Proxy logs
kubectl logs --namespace pingora-system \
  --selector app.kubernetes.io/component=proxy \
  --tail=100 --follow
```

### Check Events

```bash
kubectl get events --namespace pingora-system --sort-by='.lastTimestamp'
```

## Common Issues

### GatewayClass Not Accepted

**Symptom**: GatewayClass shows `Accepted: False`

```bash
kubectl get gatewayclass pingora
```

**Causes**:

1. Controller not running
2. Wrong controller name in GatewayClass

**Solution**:

```bash
# Verify controller is running
kubectl get pods --namespace pingora-system

# Check GatewayClass spec
kubectl get gatewayclass pingora --output yaml

# Controller name should match
controllerName: pingora.k8s.lex.la/gateway-controller
```

### Gateway Not Programmed

**Symptom**: Gateway shows `Programmed: False`

**Causes**:

1. GatewayClass not accepted
2. PingoraConfig not found or invalid
3. Unable to connect to proxy

**Solution**:

```bash
# Check GatewayClass status
kubectl get gatewayclass pingora

# Check PingoraConfig
kubectl get pingoraconfig --output wide

# Verify proxy is reachable
kubectl exec -it deployment/pingora-gateway-controller \
  --namespace pingora-system -- \
  nc -zv pingora-gateway-controller-proxy 50051
```

### HTTPRoute Not Accepted

**Symptom**: HTTPRoute shows `Accepted: False`

```bash
kubectl get httproute my-route --output yaml
```

**Causes**:

1. Gateway not found (wrong name or namespace)
2. GatewayClass not accepted
3. Backend service not found

**Solution**:

```bash
# Check parent reference
kubectl get httproute my-route --output jsonpath='{.spec.parentRefs}'

# Verify Gateway exists
kubectl get gateway --namespace pingora-system

# Check backend service
kubectl get service my-backend --namespace default
```

### Cross-Namespace Reference Failed

**Symptom**: `ResolvedRefs: False` with reason `RefNotPermitted`

**Causes**:

1. ReferenceGrant missing
2. ReferenceGrant in wrong namespace
3. ReferenceGrant doesn't match source namespace

**Solution**:

```bash
# Check for ReferenceGrant in target namespace
kubectl get referencegrant --namespace target-namespace

# Verify ReferenceGrant allows source namespace
kubectl get referencegrant allow-grant --namespace target-namespace --output yaml
```

### Routes Not Syncing

**Symptom**: Routes accepted but traffic not routing

**Causes**:

1. Proxy not receiving configuration
2. gRPC connection issues
3. Backend service has no endpoints

**Solution**:

```bash
# Check PingoraConfig status
kubectl get pingoraconfig --output yaml

# Look for sync errors in controller logs
kubectl logs --namespace pingora-system \
  --selector app.kubernetes.io/name=pingora-gateway-controller | grep -i error

# Verify backend has endpoints
kubectl get endpoints my-backend
```

### Proxy Connection Refused

**Symptom**: Controller logs show gRPC connection errors

**Causes**:

1. Proxy pods not running
2. Wrong address in PingoraConfig
3. Network policy blocking traffic

**Solution**:

```bash
# Check proxy pods
kubectl get pods --namespace pingora-system \
  --selector app.kubernetes.io/component=proxy

# Verify PingoraConfig address
kubectl get pingoraconfig --output jsonpath='{.items[*].spec.address}'

# Test connectivity
kubectl exec -it deployment/pingora-gateway-controller \
  --namespace pingora-system -- \
  nc -zv pingora-gateway-controller-proxy.pingora-system.svc.cluster.local 50051
```

### High Latency

**Symptom**: Slow request processing

**Causes**:

1. Proxy resource constraints
2. Backend service slow
3. Too many routes causing slow sync

**Solution**:

```bash
# Check proxy resource usage
kubectl top pods --namespace pingora-system \
  --selector app.kubernetes.io/component=proxy

# Check sync duration metrics
curl http://controller:8080/metrics | grep pingora_sync_duration

# Consider increasing proxy replicas/resources
```

## Debug Mode

Enable debug logging for more detailed output:

```yaml
# In Helm values
controller:
  logLevel: "debug"

proxy:
  logLevel: "debug"
```

Or patch the deployment:

```bash
kubectl set env deployment/pingora-gateway-controller \
  --namespace pingora-system \
  PINGORA_LOG_LEVEL=debug
```

## Collecting Debug Information

For bug reports, collect:

```bash
# Version information
kubectl get deployment pingora-gateway-controller \
  --namespace pingora-system \
  --output jsonpath='{.spec.template.spec.containers[0].image}'

# Resource status
kubectl get gatewayclass,gateway,httproute,grpcroute,pingoraconfig \
  --all-namespaces --output yaml > resources.yaml

# Controller logs
kubectl logs --namespace pingora-system \
  --selector app.kubernetes.io/name=pingora-gateway-controller \
  --tail=1000 > controller.log

# Proxy logs
kubectl logs --namespace pingora-system \
  --selector app.kubernetes.io/component=proxy \
  --tail=1000 > proxy.log

# Events
kubectl get events --namespace pingora-system \
  --sort-by='.lastTimestamp' > events.txt
```

## Getting Help

If issues persist:

1. Check [GitHub Issues](https://github.com/lexfrei/pingora-gateway-controller/issues)
2. Create a new issue with debug information
3. Include reproduction steps and expected behavior

## Next Steps

- Review [Metrics](metrics.md) for monitoring
- Check [Configuration](../configuration/index.md) for settings
