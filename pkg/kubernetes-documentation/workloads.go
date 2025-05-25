package kubernetesdocumentation

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WorkloadsDocumentation contains Kubernetes workloads documentation resources
type WorkloadsDocumentation struct {
	URI         string
	Name        string
	Description string
	MIMEType    string
	Content     string
}

var workloadsDocumentationResources = []WorkloadsDocumentation{
	{
		URI:         "k8s://docs/concepts/workloads/deployments",
		Name:        "Kubernetes Deployments",
		Description: "Understanding Deployments for managing Pod replicas",
		MIMEType:    "text/markdown",
		Content: `# Kubernetes Deployments

A Deployment provides declarative updates for Pods and ReplicaSets.

## What is a Deployment?

A Deployment describes a desired state for Pods and ReplicaSets, and the Deployment controller changes the actual state to the desired state at a controlled rate.

## Key Features

### Rolling Updates
- **Zero Downtime**: Updates without service interruption
- **Rollback**: Easy rollback to previous versions
- **Progressive Rollout**: Gradual replacement of old Pods

### Scaling
- **Horizontal Scaling**: Increase/decrease replica count
- **Automatic Scaling**: Using HPA (Horizontal Pod Autoscaler)

## Example Deployment YAML

` + "```yaml" + `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
` + "```" + `

## Common Commands

- Create deployment: ` + "`kubectl apply -f deployment.yaml`" + `
- Scale deployment: ` + "`kubectl scale deployment nginx-deployment --replicas=5`" + `
- Update image: ` + "`kubectl set image deployment/nginx-deployment nginx=nginx:1.22`" + `
- View rollout status: ` + "`kubectl rollout status deployment/nginx-deployment`" + `
- Rollback: ` + "`kubectl rollout undo deployment/nginx-deployment`" + `

## Deployment Strategies

1. **Rolling Update** (default): Gradual replacement
2. **Recreate**: Terminate all, then create new
3. **Blue-Green**: External tool required
4. **Canary**: Gradual traffic shifting
`,
	},
	{
		URI:         "k8s://docs/concepts/services-networking/service",
		Name:        "Kubernetes Services",
		Description: "Understanding Services for network access to Pods",
		MIMEType:    "text/markdown",
		Content: `# Kubernetes Services

A Service is an abstract way to expose an application running on a set of Pods as a network service.

## What is a Service?

With Kubernetes you don't need to modify your application to use an unfamiliar service discovery mechanism. Kubernetes gives Pods their own IP addresses and a single DNS name for a set of Pods, and can load-balance across them.

## Service Types

### ClusterIP (Default)
- **Internal Access**: Only accessible within cluster
- **Use Case**: Internal microservice communication

### NodePort
- **External Access**: Accessible via Node IP:Port
- **Port Range**: 30000-32767
- **Use Case**: Development, testing

### LoadBalancer
- **Cloud Integration**: Uses cloud provider's load balancer
- **External IP**: Gets external IP address
- **Use Case**: Production external access

### ExternalName
- **DNS Mapping**: Maps to external DNS name
- **No Proxying**: Returns CNAME record
- **Use Case**: External service integration

## Example Service YAML

` + "```yaml" + `
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
  type: ClusterIP
` + "```" + `

## Service Discovery

### DNS
- **Cluster DNS**: Automatic DNS entries
- **Format**: ` + "`<service-name>.<namespace>.svc.cluster.local`" + `

### Environment Variables
- **Automatic**: Injected into Pods
- **Format**: ` + "`<SERVICE_NAME>_SERVICE_HOST`" + `

## Common Commands

- Create service: ` + "`kubectl apply -f service.yaml`" + `
- List services: ` + "`kubectl get services`" + `
- Describe service: ` + "`kubectl describe service <service-name>`" + `
- Expose deployment: ` + "`kubectl expose deployment nginx-deployment --port=80 --type=LoadBalancer`" + `
`,
	},
}

// InitWorkloadsDocumentation initializes workloads documentation resources
func InitWorkloadsDocumentation(mcpServer *server.MCPServer) {
	for _, doc := range workloadsDocumentationResources {
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
