package kubernetesdocumentation

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ConfigMapDocumentation contains Kubernetes ConfigMap documentation resources
type ConfigMapDocumentation struct {
	URI         string
	Name        string
	Description string
	MIMEType    string
	Content     string
}

var configMapDocumentationResources = []ConfigMapDocumentation{
	{
		URI:         "k8s://docs/concepts/configuration/configmap",
		Name:        "Kubernetes ConfigMaps",
		Description: "Managing configuration data with ConfigMaps",
		MIMEType:    "text/markdown",
		Content: `# Kubernetes ConfigMaps

A ConfigMap is an API object used to store non-confidential data in key-value pairs. Pods can consume ConfigMaps as environment variables, command-line arguments, or as configuration files in a volume.

## What is a ConfigMap?

ConfigMaps allow you to decouple configuration artifacts from image content to keep containerized applications portable.

## Use Cases

1. **Environment Variables**: Inject config as env vars
2. **Command Arguments**: Pass config as command line args
3. **Configuration Files**: Mount config files as volumes
4. **Application Properties**: Store app configuration

## Creating ConfigMaps

### From Literal Values
` + "```bash" + `
kubectl create configmap app-config --from-literal=database_url=postgresql://localhost:5432/mydb
` + "```" + `

### From Files
` + "```bash" + `
kubectl create configmap app-config --from-file=config.properties
` + "```" + `

### From YAML
` + "```yaml" + `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database_url: "postgresql://localhost:5432/mydb"
  log_level: "info"
  config.properties: |
    database.host=localhost
    database.port=5432
    database.name=mydb
` + "```" + `

## Consuming ConfigMaps

### As Environment Variables
` + "```yaml" + `
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    env:
    - name: DATABASE_URL
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: database_url
` + "```" + `

### As Volume Mounts
` + "```yaml" + `
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    volumeMounts:
    - name: config-volume
      mountPath: /etc/config
  volumes:
  - name: config-volume
    configMap:
      name: app-config
` + "```" + `

## Best Practices

1. **Immutable ConfigMaps**: Set immutable: true for better performance
2. **Size Limits**: Keep under 1MB per ConfigMap
3. **Separate Concerns**: Different ConfigMaps for different purposes
4. **Versioning**: Use labels for versioning

## Common Commands

- Create ConfigMap: ` + "`kubectl create configmap <name> --from-literal=key=value`" + `
- List ConfigMaps: ` + "`kubectl get configmaps`" + `
- Describe ConfigMap: ` + "`kubectl describe configmap <name>`" + `
- Edit ConfigMap: ` + "`kubectl edit configmap <name>`" + `
- Delete ConfigMap: ` + "`kubectl delete configmap <name>`" + `
`,
	},
	{
		URI:         "k8s://docs/concepts/configuration/secret",
		Name:        "Kubernetes Secrets",
		Description: "Managing sensitive data with Secrets",
		MIMEType:    "text/markdown",
		Content: `# Kubernetes Secrets

A Secret is an object that contains a small amount of sensitive data such as a password, a token, or a key.

## What is a Secret?

Secrets are similar to ConfigMaps but are specifically intended to hold confidential data. By default, Secrets are stored unencrypted in the API server's underlying data store (etcd).

## Secret Types

### Opaque (default)
- **Arbitrary Data**: User-defined data
- **Base64 Encoded**: Data is base64 encoded

### kubernetes.io/service-account-token
- **Service Account**: For service account tokens
- **Auto-mounted**: Automatically mounted to Pods

### kubernetes.io/dockercfg
- **Docker Registry**: For Docker registry authentication
- **Legacy Format**: Old Docker config format

### kubernetes.io/dockerconfigjson
- **Docker Registry**: For Docker registry authentication
- **Modern Format**: New Docker config format

### kubernetes.io/tls
- **TLS Certificates**: For TLS certificates and keys
- **Required Keys**: tls.crt and tls.key

## Creating Secrets

### From Command Line
` + "```bash" + `
kubectl create secret generic db-secret --from-literal=username=admin --from-literal=password=secret123
` + "```" + `

### From Files
` + "```bash" + `
kubectl create secret generic ssl-secret --from-file=tls.crt --from-file=tls.key
` + "```" + `

### From YAML
` + "```yaml" + `
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  username: YWRtaW4=  # base64 encoded 'admin'
  password: c2VjcmV0MTIz  # base64 encoded 'secret123'
` + "```" + `

## Using Secrets

### As Environment Variables
` + "```yaml" + `
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    env:
    - name: DB_USERNAME
      valueFrom:
        secretKeyRef:
          name: db-secret
          key: username
    - name: DB_PASSWORD
      valueFrom:
        secretKeyRef:
          name: db-secret
          key: password
` + "```" + `

### As Volume Mounts
` + "```yaml" + `
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    volumeMounts:
    - name: secret-volume
      mountPath: /etc/secrets
      readOnly: true
  volumes:
  - name: secret-volume
    secret:
      secretName: db-secret
` + "```" + `

## Security Best Practices

1. **Enable Encryption at Rest**: Encrypt etcd data
2. **RBAC**: Use Role-Based Access Control
3. **Least Privilege**: Minimal secret access
4. **Rotation**: Regular secret rotation
5. **External Secrets**: Use external secret management systems
6. **Immutable Secrets**: Set immutable: true when possible
`,
	},
}

// InitConfigMapDocumentation initializes ConfigMap and Secret documentation resources
func InitConfigMapDocumentation(mcpServer *server.MCPServer) {
	for _, doc := range configMapDocumentationResources {
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
