package kubernetesdocumentation

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// PodDocumentation contains Kubernetes Pod documentation resources
type PodDocumentation struct {
	URI         string
	Name        string
	Description string
	MIMEType    string
	Content     string
}

var podDocumentationResources = []PodDocumentation{
	{
		URI:         "k8s://docs/concepts/workloads/pods",
		Name:        "Kubernetes Pods Concepts",
		Description: "Understanding Kubernetes Pods - the smallest deployable units",
		MIMEType:    "text/markdown",
		Content: `# Kubernetes Pods

Pods are the smallest deployable units of computing that you can create and manage in Kubernetes.

## What is a Pod?

A Pod (as in a pod of whales or pea pod) is a group of one or more containers, with shared storage and network resources, and a specification for how to run the containers.

## Key Concepts

### Pod Characteristics
- **Shared Storage**: Containers in a Pod share volumes
- **Shared Network**: Containers share the Pod's IP address and port space
- **Lifecycle**: Pods are ephemeral - they come and go
- **Single Node**: A Pod always runs on a single node

### Pod Patterns
1. **One Container per Pod**: Most common pattern
2. **Sidecar Pattern**: Helper containers alongside main container
3. **Multi-container Pod**: Multiple tightly coupled containers

## Example Pod YAML

` + "```yaml" + `
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.21
    ports:
    - containerPort: 80
` + "```" + `

## Common Commands

- Create a pod: ` + "`kubectl apply -f pod.yaml`" + `
- List pods: ` + "`kubectl get pods`" + `
- Describe pod: ` + "`kubectl describe pod <pod-name>`" + `
- Delete pod: ` + "`kubectl delete pod <pod-name>`" + `

## Best Practices

1. **Use Deployments**: Don't create Pods directly in production
2. **Resource Limits**: Always set resource requests and limits
3. **Health Checks**: Implement liveness and readiness probes
4. **Labels**: Use meaningful labels for organization

## Troubleshooting

### Common Pod Issues
- **ImagePullBackOff**: Cannot pull container image
- **CrashLoopBackOff**: Container keeps crashing
- **Pending**: Pod cannot be scheduled
- **Unknown**: Node communication issues

### Debugging Commands
` + "```bash" + `
kubectl describe pod <pod-name>
kubectl logs <pod-name>
kubectl logs <pod-name> -c <container-name>
kubectl exec -it <pod-name> -- /bin/bash
` + "```" + `
`,
	},
	{
		URI:         "k8s://docs/concepts/workloads/pods/lifecycle",
		Name:        "Pod Lifecycle",
		Description: "Understanding Pod lifecycle phases and conditions",
		MIMEType:    "text/markdown",
		Content: `# Pod Lifecycle

Understanding the lifecycle of Pods is crucial for managing applications in Kubernetes.

## Pod Phases

### Pending
- Pod has been accepted by the system
- One or more containers have not been created yet
- Includes time spent being scheduled and downloading images

### Running
- Pod has been bound to a node
- All containers have been created
- At least one container is running, starting, or restarting

### Succeeded
- All containers have terminated successfully
- Will not be restarted

### Failed
- All containers have terminated
- At least one container terminated in failure

### Unknown
- State of Pod cannot be obtained
- Typically due to communication error with the node

## Container States

### Waiting
- Container is not running
- Performing operations like pulling images or applying secrets

### Running
- Container is executing without issues

### Terminated
- Container ran to completion or failed for some reason

## Probe Types

### Liveness Probe
- Determines when to restart a container
- Kubernetes kills the container if probe fails

### Readiness Probe
- Determines when a container is ready to accept traffic
- Pod is removed from service endpoints if probe fails

### Startup Probe
- Determines when a container application has started
- Disables liveness and readiness checks until it succeeds

## Example with Probes

` + "```yaml" + `
apiVersion: v1
kind: Pod
metadata:
  name: pod-with-probes
spec:
  containers:
  - name: app
    image: myapp:latest
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
    startupProbe:
      httpGet:
        path: /startup
        port: 8080
      failureThreshold: 30
      periodSeconds: 10
` + "```" + `
`,
	},
}

// InitPodDocumentation initializes Pod documentation resources
func InitPodDocumentation(mcpServer *server.MCPServer) {
	for _, doc := range podDocumentationResources {
		resource := mcp.Resource{
			URI:         doc.URI,
			Name:        doc.Name,
			Description: doc.Description,
			MIMEType:    doc.MIMEType,
		}

		// Create handler for this specific resource
		handler := func(docContent string, docMimeType string) func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				return []mcp.ResourceContents{
					mcp.TextResourceContents{
						URI:      request.Params.URI,
						MIMEType: docMimeType,
						Text:     docContent,
					},
				}, nil
			}
		}(doc.Content, doc.MIMEType)

		mcpServer.AddResource(resource, handler)
	}
}
