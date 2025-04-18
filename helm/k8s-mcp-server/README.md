# Kubernetes MCP Server Helm Chart

This Helm chart deploys the Kubernetes MCP Server in a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

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

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `namespace.name` | Namespace in which to deploy the resources | `scoutflo-k8s-mcp` |
| `namespace.create` | Whether to create the namespace | `true` |
| `image.repository` | Image repository | `sanskardevops/kubernetes_mcp_server` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `Always` |
| `replicaCount` | Number of replicas | `1` |
| `service.type` | Service type | `ClusterIP` |
| `service.httpPort` | HTTP port | `80` |
| `service.httpsPort` | HTTPS port | `443` |
| `service.targetPort` | Target port | `8081` |
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hostname` | Hostname for ingress | `mcp.scoutflo.io` |
| `ingress.tls.enabled` | Enable TLS | `true` |
| `ingress.tls.secretName` | TLS secret name | `scoutflo-tls-cert` |
| `openai.apiKey` | OpenAI API key | `""` |
| `openai.endpoint` | OpenAI endpoint | `""` |
| `openai.deploymentName` | OpenAI deployment name | `""` |
| `openai.apiVersion` | OpenAI API version | `""` |
| `rbac.create` | Create RBAC resources | `true` |
| `dockerConfig.enabled` | Enable Docker config secret | `false` |
| `dockerConfig.dockerConfigJson` | Base64 encoded Docker config.json | `""` |
| `networkPolicy.enabled` | Enable network policy | `true` |
| `networkPolicy.namespace` | Network policy namespace | `kube-system` |

## Example installation with custom values

Create a `values.yaml` file:

```yaml
namespace:
  name: k8s-mcp

image:
  repository: sanskardevops/kubernetes_mcp_server
  tag: prod-v1.0

ingress:
  hostname: mcp.example.com
  tls:
    secretName: example-tls-cert

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