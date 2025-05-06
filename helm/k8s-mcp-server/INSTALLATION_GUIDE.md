# Kubernetes MCP Server Installation Guide

This guide provides instructions for installing the Kubernetes MCP Server Helm chart and troubleshooting common issues.

## Installation Methods

### Installing from GitHub Container Registry (Recommended)

```bash
# Install directly from GitHub Container Registry
helm install k8s-mcp-server oci://ghcr.io/scoutflo/kubernetes-mcp-server/kubernetes-mcp-server --namespace k8s-mcp-server --create-namespace
```

### Updating the Chart

```bash
helm upgrade k8s-mcp-server oci://ghcr.io/scoutflo/kubernetes-mcp-server/kubernetes-mcp-server --namespace k8s-mcp-server --version 0.1.x
```

### Installing from Local Chart

Alternatively, you can install from the local chart:

```bash
helm install kubernetes-mcp-server ./helm/k8s-mcp-server --namespace k8s-mcp-server --create-namespace
```

## Latest Released Version

The latest stable version is updated automatically through CI/CD pipelines with format `0.1.x`.

## Customizing Installation

To customize the installation, create a values file:

```yaml
# custom-values.yaml
replicaCount: 2
image:
  repository: scoutflo/kubernetes_mcp_server
  tag: latest
resources:
  limits:
    cpu: 500m
    memory: 512Mi
```

Then install with:

```bash
helm install kubernetes-mcp-server oci://ghcr.io/scoutflo/kubernetes-mcp-server -f custom-values.yaml -n k8s-mcp-server --create-namespace
```

## Updating and Pushing the Chart

### For Chart Maintainers

To update and push a new version of the chart to the OCI registry:

1. Update the chart version in `Chart.yaml`:

```yaml
version: 0.1.x  # Increment version
```

2. Package the chart:

```bash
helm package ./helm/k8s-mcp-server
```

3. Login to GitHub Container Registry:

```bash
helm registry login ghcr.io -u <username>
```

4. Push the new version:

```bash
helm push kubernetes-mcp-server-0.1.x.tgz oci://ghcr.io/scoutflo/kubernetes-mcp-server
```

### Updating Container Images

To update the container images used by the chart:

1. Edit the `values.yaml` file to update image references:

```yaml
image:
  repository: scoutflo/kubernetes_mcp_server
  tag: "production-xxx"
```

2. Package and push as described above.

## Troubleshooting

If you encounter issues during installation:

1. Check pod status: `kubectl get pods -n k8s-mcp-server`
2. View pod logs: `kubectl logs -n k8s-mcp-server <pod-name>`
3. Describe the resources: `kubectl describe deployment -n k8s-mcp-server kubernetes-mcp-server`
