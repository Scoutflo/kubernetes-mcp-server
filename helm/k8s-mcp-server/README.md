# Kubernetes MCP Server Helm Chart

This Helm chart deploys the Kubernetes MCP Server in a Kubernetes cluster with production-ready configurations.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Prometheus Operator (optional, for ServiceMonitor)

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
# Add the repository (if needed)
# helm repo add scoutflo https://scoutflo.github.io/charts

# Update your repositories (if needed)
# helm repo update

# Install the chart
helm install my-release ./helm/k8s-mcp-server
```

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
helm uninstall my-release
```

## Production-Ready Features

This Helm chart includes several production-ready features:

1. **Pod Disruption Budget** - Ensures high availability during voluntary disruptions
2. **Liveness and Readiness Probes** - Monitors application health with `/healthz` and `/readyz` endpoints
3. **Resource Management** - Configurable CPU and memory limits/requests
4. **ServiceMonitor** - Optional Prometheus monitoring integration
5. **Metrics Endpoint** - Optional metrics exposure on port 9090
6. **Pod Annotations/Labels** - Customizable annotations and labels for pods
7. **Well-documented values.yaml** - Complete documentation for all configuration options

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `namespace.name` | Namespace in which to deploy the resources | `k8s-mcp-server` |
| `namespace.create` | Whether to create the namespace | `false` |
| `image.repository` | Image repository | `scoutflo/kubernetes_mcp_server` |
| `image.tag` | Image tag | `pre-prod-18` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `replicaCount` | Number of replicas | `1` |
| `service.type` | Service type | `ClusterIP` |
| `service.httpPort` | HTTP port | `80` |
| `service.httpsPort` | HTTPS port | `443` |
| `service.targetPort` | Target port | `8081` |
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hostname` | Hostname for ingress | `k8s-mcp-server.scoutflo.agency` |
| `ingress.tls.enabled` | Enable TLS | `true` |
| `ingress.tls.secretName` | TLS secret name | `scoutflo-agency-wildcard-tls` |
| `openai.apiKey` | OpenAI API key | `""` |
| `openai.endpoint` | OpenAI endpoint | `""` |
| `openai.deploymentName` | OpenAI deployment name | `""` |
| `openai.apiVersion` | OpenAI API version | `""` |
| `rbac.create` | Create RBAC resources | `true` |
| `resources.limits.cpu` | CPU limits | `500m` |
| `resources.limits.memory` | Memory limits | `512Mi` |
| `resources.requests.cpu` | CPU requests | `100m` |
| `resources.requests.memory` | Memory requests | `128Mi` |
| `podDisruptionBudget.enabled` | Enable Pod Disruption Budget | `true` |
| `podDisruptionBudget.minAvailable` | Minimum available pods | `1` |
| `podDisruptionBudget.maxUnavailable` | Maximum unavailable pods | `nil` |
| `probes.liveness.enabled` | Enable liveness probe | `true` |
| `probes.liveness.initialDelaySeconds` | Initial delay seconds for liveness probe | `30` |
| `probes.liveness.periodSeconds` | Period seconds for liveness probe | `10` |
| `probes.liveness.timeoutSeconds` | Timeout seconds for liveness probe | `5` |
| `probes.liveness.failureThreshold` | Failure threshold for liveness probe | `6` |
| `probes.liveness.successThreshold` | Success threshold for liveness probe | `1` |
| `probes.readiness.enabled` | Enable readiness probe | `true` |
| `probes.readiness.initialDelaySeconds` | Initial delay seconds for readiness probe | `10` |
| `probes.readiness.periodSeconds` | Period seconds for readiness probe | `10` |
| `probes.readiness.timeoutSeconds` | Timeout seconds for readiness probe | `5` |
| `probes.readiness.failureThreshold` | Failure threshold for readiness probe | `6` |
| `probes.readiness.successThreshold` | Success threshold for readiness probe | `1` |
| `metrics.enabled` | Enable metrics | `false` |
| `metrics.port` | Metrics port | `9090` |
| `metrics.serviceMonitor.enabled` | Enable ServiceMonitor | `false` |
| `metrics.serviceMonitor.namespace` | ServiceMonitor namespace | `""` |
| `metrics.serviceMonitor.interval` | ServiceMonitor scrape interval | `30s` |
| `metrics.serviceMonitor.scrapeTimeout` | ServiceMonitor scrape timeout | `10s` |
| `podAnnotations` | Additional pod annotations | `{}` |
| `podLabels` | Additional pod labels | `{}` |

## Example installation with custom values

Create a `values.yaml` file:

```yaml
namespace:
  name: k8s-mcp-server

image:
  repository: scoutflo/kubernetes_mcp_server
  tag: latest

replicaCount: 2

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

podDisruptionBudget:
  enabled: true
  minAvailable: 1

metrics:
  enabled: true
  serviceMonitor:
    enabled: true

probes:
  liveness:
    enabled: true
  readiness:
    enabled: true
    
openai:
  apiKey: your-api-key
  endpoint: https://your-endpoint.openai.azure.com/
  deploymentName: gpt-4o
  apiVersion: 2024-02-15-preview
```

Install the chart with custom values:

```bash
helm install my-release ./helm/k8s-mcp-server -f values.yaml
``` 