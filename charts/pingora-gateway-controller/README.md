# pingora-gateway-controller

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

Kubernetes Gateway API controller for Pingora proxy

**Homepage:** <https://github.com/lexfrei/pingora-gateway-controller/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| lexfrei | <f@lex.la> | <https://github.com/lexfrei> |

## Source Code

* <https://github.com/lexfrei/pingora-gateway-controller/>

## Requirements

Kubernetes: `>=1.25.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod scheduling |
| controller | object | `{"clusterDomain":"","controllerName":"pingora.k8s.lex.la/gateway-controller","gatewayClassName":"pingora","logFormat":"json","logLevel":"info"}` | Controller configuration |
| controller.clusterDomain | string | auto-detected from /etc/resolv.conf, fallback: cluster.local | Kubernetes cluster domain for service DNS resolution |
| controller.controllerName | string | `"pingora.k8s.lex.la/gateway-controller"` | Controller name for GatewayClass (must be unique in cluster) |
| controller.gatewayClassName | string | `"pingora"` | GatewayClass name to watch |
| controller.logFormat | string | `"json"` | Log format (json, text) |
| controller.logLevel | string | `"info"` | Log level (debug, info, warn, error) |
| dnsConfig | object | `{}` | Custom DNS configuration for pod |
| dnsPolicy | string | `""` | DNS policy for pod (ClusterFirst, Default, ClusterFirstWithHostNet, None) |
| fullnameOverride | string | `""` | Override the full release name |
| gatewayClass | object | `{"create":true}` | GatewayClass configuration |
| gatewayClass.create | bool | `true` | Create GatewayClass resource |
| healthProbes | object | `{"livenessProbe":{"enabled":true,"failureThreshold":3,"initialDelaySeconds":15,"periodSeconds":20,"timeoutSeconds":5},"readinessProbe":{"enabled":true,"failureThreshold":3,"initialDelaySeconds":5,"periodSeconds":10,"timeoutSeconds":3},"startupProbe":{"enabled":true,"failureThreshold":12,"initialDelaySeconds":0,"periodSeconds":5,"timeoutSeconds":3}}` | Health probes configuration |
| healthProbes.livenessProbe | object | `{"enabled":true,"failureThreshold":3,"initialDelaySeconds":15,"periodSeconds":20,"timeoutSeconds":5}` | Liveness probe configuration |
| healthProbes.readinessProbe | object | `{"enabled":true,"failureThreshold":3,"initialDelaySeconds":5,"periodSeconds":10,"timeoutSeconds":3}` | Readiness probe configuration |
| healthProbes.startupProbe | object | `{"enabled":true,"failureThreshold":12,"initialDelaySeconds":0,"periodSeconds":5,"timeoutSeconds":3}` | Startup probe configuration |
| image | object | `{"pullPolicy":"IfNotPresent","repository":"ghcr.io/lexfrei/pingora-gateway-controller","tag":""}` | Container image configuration |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"ghcr.io/lexfrei/pingora-gateway-controller"` | Image repository |
| image.tag | string | `""` | Image tag (defaults to appVersion) |
| imagePullSecrets | list | `[]` | Image pull secrets for private registries |
| leaderElection | object | `{"enabled":false,"leaseName":"pingora-gateway-controller-leader","namespace":""}` | Leader election configuration for high availability |
| leaderElection.enabled | bool | `false` | Enable leader election (required for running multiple replicas) |
| leaderElection.leaseName | string | `"pingora-gateway-controller-leader"` | Name of the leader election lease |
| leaderElection.namespace | string | `""` | Namespace for leader election lease (defaults to release namespace) |
| nameOverride | string | `""` | Override the chart name |
| networkPolicy | object | `{"enabled":false,"ingress":{"from":[]},"pingoraProxy":{"namespaceSelector":{},"podSelector":{},"port":50051}}` | NetworkPolicy configuration |
| networkPolicy.enabled | bool | `false` | Enable NetworkPolicy for controller pods |
| networkPolicy.ingress | object | `{"from":[]}` | Ingress source configuration |
| networkPolicy.ingress.from | list | `[]` | Allow ingress from specific namespaces/pods |
| networkPolicy.pingoraProxy | object | `{"namespaceSelector":{},"podSelector":{},"port":50051}` | Pingora proxy egress configuration |
| networkPolicy.pingoraProxy.namespaceSelector | object | `{}` | Namespace selector for Pingora proxy pods |
| networkPolicy.pingoraProxy.podSelector | object | `{}` | Pod selector for Pingora proxy pods |
| networkPolicy.pingoraProxy.port | int | `50051` | gRPC port for Pingora proxy |
| nodeSelector | object | `{}` | Node selector for pod scheduling |
| pingoraConfig | object | `{"address":"","connection":{"connectTimeoutSeconds":5,"keepaliveTimeSeconds":30,"maxRetries":3,"requestTimeoutSeconds":30,"retryBackoffMs":1000},"create":true,"name":"","tls":{"enabled":false,"insecureSkipVerify":false,"secretRef":{"name":"","namespace":""},"serverName":""}}` | PingoraConfig configuration Reference configuration for the Pingora proxy connection. |
| pingoraConfig.address | string | `""` | gRPC endpoint address of the Pingora proxy Format: "host:port" (e.g., "pingora-proxy.pingora-system.svc.cluster.local:50051") |
| pingoraConfig.connection | object | `{"connectTimeoutSeconds":5,"keepaliveTimeSeconds":30,"maxRetries":3,"requestTimeoutSeconds":30,"retryBackoffMs":1000}` | Connection parameters |
| pingoraConfig.connection.connectTimeoutSeconds | int | `5` | Timeout for establishing connection (seconds) |
| pingoraConfig.connection.keepaliveTimeSeconds | int | `30` | Interval for keepalive pings (seconds) |
| pingoraConfig.connection.maxRetries | int | `3` | Maximum number of retries for failed requests |
| pingoraConfig.connection.requestTimeoutSeconds | int | `30` | Timeout for individual gRPC requests (seconds) |
| pingoraConfig.connection.retryBackoffMs | int | `1000` | Backoff duration between retries (milliseconds) |
| pingoraConfig.create | bool | `true` | Create PingoraConfig resource |
| pingoraConfig.name | string | `""` | Name of the PingoraConfig (defaults to release fullname) |
| pingoraConfig.tls | object | `{"enabled":false,"insecureSkipVerify":false,"secretRef":{"name":"","namespace":""},"serverName":""}` | TLS configuration for gRPC connection |
| pingoraConfig.tls.enabled | bool | `false` | Enable TLS for gRPC connection |
| pingoraConfig.tls.insecureSkipVerify | bool | `false` | Skip TLS certificate verification (WARNING: for testing only) |
| pingoraConfig.tls.secretRef | object | `{"name":"","namespace":""}` | Reference to Secret containing TLS certificates The Secret must contain "tls.crt" and "tls.key" keys. Optionally include "ca.crt" for CA validation. |
| pingoraConfig.tls.secretRef.name | string | `""` | Name of the Secret containing TLS certificates |
| pingoraConfig.tls.secretRef.namespace | string | `""` | Namespace of the Secret (defaults to release namespace) |
| pingoraConfig.tls.serverName | string | `""` | Override server name for TLS verification |
| podAnnotations | object | `{}` | Annotations to add to pods |
| podDisruptionBudget | object | `{"enabled":false,"maxUnavailable":null,"minAvailable":1,"unhealthyPodEvictionPolicy":"IfHealthyBudget"}` | PodDisruptionBudget configuration for high availability |
| podDisruptionBudget.enabled | bool | `false` | Enable PodDisruptionBudget |
| podDisruptionBudget.maxUnavailable | string | `nil` | Maximum number of unavailable pods during disruptions |
| podDisruptionBudget.minAvailable | int | `1` | Minimum number of available pods during disruptions |
| podDisruptionBudget.unhealthyPodEvictionPolicy | string | `"IfHealthyBudget"` | Policy for evicting unhealthy pods (IfHealthyBudget, AlwaysAllow) |
| podLabels | object | `{}` | Additional labels to add to pods |
| podSecurityContext | object | See values.yaml | Pod security context (secure defaults) |
| priorityClassName | string | `""` | Priority class name for pod scheduling priority |
| proxy | object | `{"affinity":{},"enabled":true,"healthProbes":{"livenessProbe":{"failureThreshold":3,"initialDelaySeconds":15,"periodSeconds":20,"timeoutSeconds":5},"readinessProbe":{"failureThreshold":3,"initialDelaySeconds":5,"periodSeconds":10,"timeoutSeconds":3},"startupProbe":{"enabled":true,"failureThreshold":12,"initialDelaySeconds":0,"periodSeconds":5,"timeoutSeconds":3}},"image":{"pullPolicy":"IfNotPresent","repository":"ghcr.io/lexfrei/pingora-proxy","tag":""},"logLevel":"info","nodeSelector":{},"podAnnotations":{},"podLabels":{},"priorityClassName":"","replicaCount":2,"resources":{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"service":{"annotations":{},"type":"ClusterIP"},"terminationGracePeriodSeconds":30,"tolerations":[]}` | Pingora proxy deployment configuration |
| proxy.affinity | object | `{}` | Affinity rules for pod scheduling |
| proxy.enabled | bool | `true` | Enable proxy deployment |
| proxy.healthProbes | object | `{"livenessProbe":{"failureThreshold":3,"initialDelaySeconds":15,"periodSeconds":20,"timeoutSeconds":5},"readinessProbe":{"failureThreshold":3,"initialDelaySeconds":5,"periodSeconds":10,"timeoutSeconds":3},"startupProbe":{"enabled":true,"failureThreshold":12,"initialDelaySeconds":0,"periodSeconds":5,"timeoutSeconds":3}}` | Health probes configuration |
| proxy.healthProbes.livenessProbe | object | `{"failureThreshold":3,"initialDelaySeconds":15,"periodSeconds":20,"timeoutSeconds":5}` | Liveness probe configuration |
| proxy.healthProbes.readinessProbe | object | `{"failureThreshold":3,"initialDelaySeconds":5,"periodSeconds":10,"timeoutSeconds":3}` | Readiness probe configuration |
| proxy.healthProbes.startupProbe | object | `{"enabled":true,"failureThreshold":12,"initialDelaySeconds":0,"periodSeconds":5,"timeoutSeconds":3}` | Startup probe configuration |
| proxy.image | object | `{"pullPolicy":"IfNotPresent","repository":"ghcr.io/lexfrei/pingora-proxy","tag":""}` | Container image configuration |
| proxy.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| proxy.image.repository | string | `"ghcr.io/lexfrei/pingora-proxy"` | Image repository |
| proxy.image.tag | string | `""` | Image tag (defaults to appVersion) |
| proxy.logLevel | string | `"info"` | Log level for proxy (trace, debug, info, warn, error) |
| proxy.nodeSelector | object | `{}` | Node selector for pod scheduling |
| proxy.podAnnotations | object | `{}` | Annotations to add to proxy pods |
| proxy.podLabels | object | `{}` | Additional labels to add to proxy pods |
| proxy.priorityClassName | string | `""` | Priority class name for pod scheduling priority |
| proxy.replicaCount | int | `2` | Number of proxy replicas |
| proxy.resources | object | `{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Container resource requests and limits |
| proxy.service | object | `{"annotations":{},"type":"ClusterIP"}` | Service configuration |
| proxy.service.annotations | object | `{}` | Service annotations |
| proxy.service.type | string | `"ClusterIP"` | Service type |
| proxy.terminationGracePeriodSeconds | int | `30` | Termination grace period in seconds |
| proxy.tolerations | list | `[]` | Tolerations for pod scheduling |
| replicaCount | int | `1` | Number of controller replicas |
| resources | object | `{"limits":{"cpu":"200m","memory":"256Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Container resource requests and limits |
| securityContext | object | See values.yaml | Container security context (secure defaults) |
| service | object | `{"annotations":{},"healthPort":8081,"metricsPort":8080,"type":"ClusterIP"}` | Service configuration |
| service.annotations | object | `{}` | Service annotations |
| service.healthPort | int | `8081` | Health check endpoint port |
| service.metricsPort | int | `8080` | Metrics endpoint port |
| service.type | string | `"ClusterIP"` | Service type |
| serviceAccount | object | `{"annotations":{},"name":""}` | Service account configuration |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.name | string | `""` | The name of the service account to use If empty, uses the fullname template (release-name-chart-name) |
| serviceMonitor | object | `{"enabled":false,"interval":"","labels":{}}` | ServiceMonitor configuration for Prometheus Operator |
| serviceMonitor.enabled | bool | `false` | Enable ServiceMonitor creation |
| serviceMonitor.interval | string | `""` | Scrape interval (uses Prometheus default if empty) |
| serviceMonitor.labels | object | `{}` | Additional labels for ServiceMonitor (for Prometheus selector) |
| terminationGracePeriodSeconds | int | `30` | Termination grace period in seconds for graceful shutdown |
| tolerations | list | `[]` | Tolerations for pod scheduling |
| topologySpreadConstraints | list | `[]` | Topology spread constraints for pod distribution |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
