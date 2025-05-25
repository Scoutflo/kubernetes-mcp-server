package kubernetesdocumentation

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// KubectlDocumentation contains kubectl command reference resources
type KubectlDocumentation struct {
	URI         string
	Name        string
	Description string
	MIMEType    string
	Content     string
}

var kubectlDocumentationResources = []KubectlDocumentation{
	{
		URI:         "k8s://docs/reference/kubectl/commands",
		Name:        "Kubectl Command Reference",
		Description: "Common kubectl commands and usage patterns",
		MIMEType:    "text/markdown",
		Content: `# Kubectl Command Reference

kubectl is the command-line interface for running commands against Kubernetes clusters.

## Basic Commands

### Cluster Information
` + "```bash" + `
kubectl cluster-info                    # Display cluster info
kubectl version                         # Show client and server versions
kubectl api-resources                   # List available API resources
kubectl api-versions                    # List available API versions
` + "```" + `

### Resource Management
` + "```bash" + `
kubectl get <resource>                  # List resources
kubectl describe <resource> <name>      # Show detailed information
kubectl create -f <file>                # Create resource from file
kubectl apply -f <file>                 # Apply configuration
kubectl delete <resource> <name>        # Delete resource
kubectl edit <resource> <name>          # Edit resource
` + "```" + `

## Working with Pods

### Basic Pod Operations
` + "```bash" + `
kubectl get pods                        # List all pods
kubectl get pods -o wide                # List pods with additional info
kubectl get pods --all-namespaces       # List pods in all namespaces
kubectl get pods -l app=nginx           # List pods with label selector
kubectl describe pod <pod-name>         # Detailed pod information
kubectl logs <pod-name>                 # View pod logs
kubectl logs -f <pod-name>              # Follow log output
kubectl exec -it <pod-name> -- /bin/bash # Execute command in pod
` + "```" + `

### Pod Troubleshooting
` + "```bash" + `
kubectl get events                      # View cluster events
kubectl get events --sort-by=.metadata.creationTimestamp
kubectl top pods                        # Show resource usage
kubectl port-forward <pod-name> 8080:80 # Forward local port to pod
` + "```" + `

## Working with Services

### Service Operations
` + "```bash" + `
kubectl get services                    # List services
kubectl get svc                         # Shorthand for services
kubectl describe service <service-name> # Service details
kubectl expose deployment <name> --port=80 --type=LoadBalancer
kubectl port-forward service/<service-name> 8080:80
` + "```" + `

## Working with Deployments

### Deployment Operations
` + "```bash" + `
kubectl get deployments                 # List deployments
kubectl describe deployment <name>      # Deployment details
kubectl scale deployment <name> --replicas=5
kubectl rollout status deployment <name>
kubectl rollout history deployment <name>
kubectl rollout undo deployment <name>
kubectl set image deployment/<name> container=image:tag
` + "```" + `

## Configuration and Context

### Context Management
` + "```bash" + `
kubectl config view                     # Show kubeconfig
kubectl config get-contexts            # List available contexts
kubectl config current-context         # Show current context
kubectl config use-context <context>   # Switch context
kubectl config set-context <context> --namespace=<namespace>
` + "```" + `

### Namespace Operations
` + "```bash" + `
kubectl get namespaces                  # List namespaces
kubectl create namespace <name>         # Create namespace
kubectl delete namespace <name>         # Delete namespace
kubectl config set-context --current --namespace=<namespace>
` + "```" + `

## Advanced Commands

### Resource Management
` + "```bash" + `
kubectl apply -f directory/             # Apply all files in directory
kubectl apply -k kustomization/         # Apply Kustomize configuration
kubectl diff -f file.yaml               # Show differences
kubectl dry-run=client -f file.yaml     # Validate without creating
` + "```" + `

### Debugging and Troubleshooting
` + "```bash" + `
kubectl get events --sort-by=.metadata.creationTimestamp
kubectl describe nodes                  # Node information
kubectl top nodes                       # Node resource usage
kubectl get pods --field-selector=status.phase=Failed
kubectl get all                         # List common resources
` + "```" + `

### Output Formatting
` + "```bash" + `
kubectl get pods -o yaml               # YAML output
kubectl get pods -o json               # JSON output
kubectl get pods -o wide               # Additional columns
kubectl get pods -o custom-columns=NAME:.metadata.name,STATUS:.status.phase
kubectl get pods --show-labels         # Show labels
` + "```" + `

## Useful Shortcuts

### Resource Abbreviations
- pods = po
- services = svc
- deployments = deploy
- replicasets = rs
- namespaces = ns
- nodes = no
- persistentvolumes = pv
- persistentvolumeclaims = pvc
- configmaps = cm
- secrets = secret

### Common Patterns
` + "```bash" + `
kubectl get po -A                      # All pods in all namespaces
kubectl get all -n <namespace>         # All resources in namespace
kubectl delete pod <name> --force --grace-period=0  # Force delete
kubectl run nginx --image=nginx --dry-run=client -o yaml > pod.yaml
` + "```" + `
`,
	},
}

// InitKubectlDocumentation initializes kubectl command reference resources
func InitKubectlDocumentation(mcpServer *server.MCPServer) {
	for _, doc := range kubectlDocumentationResources {
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
