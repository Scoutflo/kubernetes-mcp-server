# Health Check Package

This package provides HTTP health check endpoints for the Kubernetes MCP Server when running in SSE mode.

## Endpoints

- `/healthz` - Liveness probe endpoint
  - Returns 200 OK if the server is running
  - Used by Kubernetes to determine if the pod should be restarted

- `/readyz` - Readiness probe endpoint
  - Returns 200 OK if the server is ready to receive requests
  - Returns 503 Service Unavailable if the server is not ready
  - Used by Kubernetes to determine if traffic should be sent to the pod

## Usage

These endpoints are automatically added to a separate HTTP server when running in SSE mode. The health check server runs on a configurable port (default: 8082).

## Helm Configuration

These endpoints can be used with the Kubernetes liveness and readiness probes in the Helm chart. The health check port is configurable via the `service.healthPort` value:

```yaml
service:
  healthPort: 8082  # Port for health checks
```

The probes are configured as follows:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: {{ .Values.service.healthPort }}
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 6
  successThreshold: 1

readinessProbe:
  httpGet:
    path: /readyz
    port: {{ .Values.service.healthPort }}
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 6
  successThreshold: 1
```

To enable the probes in the Helm chart, set the following in your values.yaml:

```yaml
probes:
  liveness:
    enabled: true
    # other parameters...
  readiness:
    enabled: true
    # other parameters...
``` 