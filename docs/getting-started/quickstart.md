# Quick Start

This guide walks you through creating your first HTTPRoute to expose a
Kubernetes service through Pingora proxy.

## Prerequisites

Ensure you have completed:

- [Prerequisites](prerequisites.md) - Gateway API CRDs installed
- [Installation](installation.md) - Controller installed and running

## Deploy a Sample Application

First, deploy a simple application to expose:

```bash
kubectl create deployment nginx --image=nginx:latest
kubectl expose deployment nginx --port=80
```

## Create a Gateway

Create a Gateway resource to define an entry point:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: pingora-gateway
  namespace: pingora-system
spec:
  gatewayClassName: pingora
  listeners:
    - name: http
      port: 80
      protocol: HTTP
```

Apply the Gateway:

```bash
kubectl apply --filename gateway.yaml
```

## Create an HTTPRoute

Create an HTTPRoute to expose the nginx service:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: nginx
  namespace: default
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - nginx.example.com
  rules:
    - backendRefs:
        - name: nginx
          port: 80
```

Apply the route:

```bash
kubectl apply --filename httproute.yaml
```

## Verify the Route

Check that the HTTPRoute is accepted:

```bash
kubectl get httproute nginx
```

Expected output:

```text
NAME    HOSTNAMES               AGE
nginx   ["nginx.example.com"]   30s
```

Check the route status:

```bash
kubectl get httproute nginx --output jsonpath='{.status.parents[*].conditions}' | jq
```

Expected output includes `"type":"Accepted","status":"True"`.

## Access Your Application

Get the proxy service IP:

```bash
kubectl get service --namespace pingora-system --selector app.kubernetes.io/component=proxy
```

For testing, use port-forward:

```bash
kubectl port-forward --namespace pingora-system service/pingora-gateway-controller-proxy 8080:80
```

Then access via curl:

```bash
curl --header "Host: nginx.example.com" http://localhost:8080
```

## Path-Based Routing

Route different paths to different services:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: api-routes
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - api.example.com
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /v1
      backendRefs:
        - name: api-v1
          port: 8080
    - matches:
        - path:
            type: PathPrefix
            value: /v2
      backendRefs:
        - name: api-v2
          port: 8080
```

## Header-Based Routing

Route based on request headers:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: header-routes
spec:
  parentRefs:
    - name: pingora-gateway
      namespace: pingora-system
  hostnames:
    - app.example.com
  rules:
    - matches:
        - headers:
            - name: X-Version
              value: beta
      backendRefs:
        - name: app-beta
          port: 8080
    - backendRefs:
        - name: app-stable
          port: 8080
```

## Troubleshooting

### Route Not Accepted

Check controller logs:

```bash
kubectl logs --selector app.kubernetes.io/name=pingora-gateway-controller \
  --namespace pingora-system
```

Common issues:

- Gateway not found (wrong namespace or name in parentRefs)
- Service not found (wrong service name or namespace)
- GatewayClass not accepted

### Connection Refused

Verify the proxy is running and has endpoints:

```bash
kubectl get endpoints --namespace pingora-system
```

Check if backend service has healthy pods:

```bash
kubectl get pods --selector app=nginx
```

## Next Steps

- Learn about [Configuration](../configuration/index.md) options
- Explore [Gateway API](../gateway-api/index.md) features
- Set up [Monitoring](../guides/monitoring.md) for production
