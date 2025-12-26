# Manual Installation

Step-by-step guide for installing Pingora Gateway Controller without Helm.

## Prerequisites

- Kubernetes 1.25+
- kubectl configured
- Gateway API CRDs installed

## Install Gateway API CRDs

```bash
kubectl apply --filename https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/standard-install.yaml
```

## Install Controller CRDs

Apply the PingoraConfig CRD:

```bash
kubectl apply --filename https://raw.githubusercontent.com/lexfrei/pingora-gateway-controller/master/charts/pingora-gateway-controller/crds/pingoraconfig-crd.yaml
```

## Create Namespace

```bash
kubectl create namespace pingora-system
```

## Create RBAC Resources

### Service Account

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pingora-gateway-controller
  namespace: pingora-system
```

### ClusterRole

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pingora-gateway-controller
rules:
  # Gateway API resources
  - apiGroups: ["gateway.networking.k8s.io"]
    resources: ["gatewayclasses", "gateways", "httproutes", "grpcroutes", "referencegrants"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["gateway.networking.k8s.io"]
    resources: ["gatewayclasses/status", "gateways/status", "httproutes/status", "grpcroutes/status"]
    verbs: ["update", "patch"]

  # PingoraConfig CRD
  - apiGroups: ["pingora.k8s.lex.la"]
    resources: ["pingoraconfigs"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["pingora.k8s.lex.la"]
    resources: ["pingoraconfigs/status"]
    verbs: ["update", "patch"]

  # Core resources
  - apiGroups: [""]
    resources: ["services", "endpoints", "secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

  # Leader election
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "create", "update"]
```

### ClusterRoleBinding

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: pingora-gateway-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pingora-gateway-controller
subjects:
  - kind: ServiceAccount
    name: pingora-gateway-controller
    namespace: pingora-system
```

Apply RBAC:

```bash
kubectl apply --filename rbac.yaml
```

## Deploy Controller

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pingora-gateway-controller
  namespace: pingora-system
  labels:
    app.kubernetes.io/name: pingora-gateway-controller
    app.kubernetes.io/component: controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: pingora-gateway-controller
  template:
    metadata:
      labels:
        app.kubernetes.io/name: pingora-gateway-controller
        app.kubernetes.io/component: controller
    spec:
      serviceAccountName: pingora-gateway-controller
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: controller
          image: ghcr.io/lexfrei/pingora-gateway-controller:latest
          args:
            - --gateway-class-name=pingora
            - --controller-name=pingora.k8s.lex.la/gateway-controller
            - --metrics-addr=:8080
            - --health-addr=:8081
            - --log-level=info
            - --log-format=json
          ports:
            - name: metrics
              containerPort: 8080
            - name: health
              containerPort: 8081
          livenessProbe:
            httpGet:
              path: /healthz
              port: health
            initialDelaySeconds: 15
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /readyz
              port: health
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            limits:
              cpu: 200m
              memory: 256Mi
            requests:
              cpu: 100m
              memory: 128Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
```

Apply controller:

```bash
kubectl apply --filename controller-deployment.yaml
```

## Deploy Proxy

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pingora-gateway-controller-proxy
  namespace: pingora-system
  labels:
    app.kubernetes.io/name: pingora-gateway-controller
    app.kubernetes.io/component: proxy
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: pingora-gateway-controller
      app.kubernetes.io/component: proxy
  template:
    metadata:
      labels:
        app.kubernetes.io/name: pingora-gateway-controller
        app.kubernetes.io/component: proxy
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: proxy
          image: ghcr.io/lexfrei/pingora-proxy:latest
          args:
            - --log-level=info
          ports:
            - name: http
              containerPort: 80
            - name: grpc
              containerPort: 50051
            - name: health
              containerPort: 8081
          livenessProbe:
            httpGet:
              path: /health
              port: health
            initialDelaySeconds: 15
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /health
              port: health
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            limits:
              cpu: 500m
              memory: 512Mi
            requests:
              cpu: 100m
              memory: 128Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
---
apiVersion: v1
kind: Service
metadata:
  name: pingora-gateway-controller-proxy
  namespace: pingora-system
spec:
  selector:
    app.kubernetes.io/name: pingora-gateway-controller
    app.kubernetes.io/component: proxy
  ports:
    - name: http
      port: 80
      targetPort: 80
    - name: grpc
      port: 50051
      targetPort: 50051
```

Apply proxy:

```bash
kubectl apply --filename proxy-deployment.yaml
```

## Create GatewayClass

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: pingora
spec:
  controllerName: pingora.k8s.lex.la/gateway-controller
  parametersRef:
    group: pingora.k8s.lex.la
    kind: PingoraConfig
    name: pingora-config
```

## Create PingoraConfig

```yaml
apiVersion: pingora.k8s.lex.la/v1alpha1
kind: PingoraConfig
metadata:
  name: pingora-config
spec:
  address: "pingora-gateway-controller-proxy.pingora-system.svc.cluster.local:50051"
  connection:
    connectTimeoutSeconds: 5
    requestTimeoutSeconds: 30
    maxRetries: 3
```

Apply configuration:

```bash
kubectl apply --filename gatewayclass.yaml
kubectl apply --filename pingoraconfig.yaml
```

## Create Gateway

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

Apply Gateway:

```bash
kubectl apply --filename gateway.yaml
```

## Verify Installation

```bash
# Check pods
kubectl get pods --namespace pingora-system

# Check GatewayClass
kubectl get gatewayclass pingora

# Check Gateway
kubectl get gateway --namespace pingora-system

# Check PingoraConfig
kubectl get pingoraconfig
```

## Create Controller Service (for metrics)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: pingora-gateway-controller
  namespace: pingora-system
spec:
  selector:
    app.kubernetes.io/name: pingora-gateway-controller
    app.kubernetes.io/component: controller
  ports:
    - name: metrics
      port: 8080
    - name: health
      port: 8081
```

## Cleanup

To remove all resources:

```bash
kubectl delete namespace pingora-system
kubectl delete clusterrole pingora-gateway-controller
kubectl delete clusterrolebinding pingora-gateway-controller
kubectl delete gatewayclass pingora
kubectl delete pingoraconfig pingora-config
```

## Next Steps

- Create [HTTPRoute](../gateway-api/httproute.md) to route traffic
- Set up [Monitoring](../guides/monitoring.md) for production
