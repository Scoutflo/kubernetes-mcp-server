package kubernetesdocumentation

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StorageDocumentation contains Kubernetes storage documentation resources
type StorageDocumentation struct {
	URI         string
	Name        string
	Description string
	MIMEType    string
	Content     string
}

var storageDocumentationResources = []StorageDocumentation{
	{
		URI:         "k8s://docs/concepts/storage/persistent-volumes",
		Name:        "Kubernetes Persistent Volumes",
		Description: "Understanding Persistent Volumes and storage in Kubernetes",
		MIMEType:    "text/markdown",
		Content: `# Kubernetes Persistent Volumes

Persistent Volumes (PV) are a way to provide durable storage in a Kubernetes cluster. They are resources in the cluster just like nodes are cluster resources.

## Storage Concepts

### Persistent Volume (PV)
- **Cluster Resource**: Provisioned by admin or dynamically
- **Lifecycle**: Independent of Pod lifecycle
- **Storage Class**: Defines storage properties

### Persistent Volume Claim (PVC)
- **User Request**: Request for storage by a user
- **Binding**: Binds to available PV
- **Resource Requirements**: Size, access modes, storage class

### Storage Class
- **Dynamic Provisioning**: Automatically creates PVs
- **Storage Provider**: Defines which storage to use
- **Parameters**: Storage-specific parameters

## Volume Access Modes

1. **ReadWriteOnce (RWO)**: Mount as read-write by single node
2. **ReadOnlyMany (ROX)**: Mount as read-only by many nodes
3. **ReadWriteMany (RWX)**: Mount as read-write by many nodes
4. **ReadWriteOncePod (RWOP)**: Mount as read-write by single pod

## Example PV and PVC

### Persistent Volume
` + "```yaml" + `
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-example
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  hostPath:
    path: /mnt/data
` + "```" + `

### Persistent Volume Claim
` + "```yaml" + `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-example
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: manual
` + "```" + `

### Using PVC in Pod
` + "```yaml" + `
apiVersion: v1
kind: Pod
metadata:
  name: pod-with-pvc
spec:
  containers:
  - name: app
    image: nginx
    volumeMounts:
    - name: storage
      mountPath: /usr/share/nginx/html
  volumes:
  - name: storage
    persistentVolumeClaim:
      claimName: pvc-example
` + "```" + `

## Storage Classes

### Dynamic Provisioning Example
` + "```yaml" + `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
  fsType: ext4
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
` + "```" + `

## Reclaim Policies

1. **Retain**: Manual reclamation required
2. **Delete**: Automatic deletion when PVC deleted
3. **Recycle**: Deprecated, use dynamic provisioning

## Common Commands

- List PVs: ` + "`kubectl get pv`" + `
- List PVCs: ` + "`kubectl get pvc`" + `
- List Storage Classes: ` + "`kubectl get storageclass`" + `
- Describe PV: ` + "`kubectl describe pv <pv-name>`" + `
- Describe PVC: ` + "`kubectl describe pvc <pvc-name>`" + `

## Best Practices

1. **Use Storage Classes**: Enable dynamic provisioning
2. **Size Appropriately**: Right-size storage requests
3. **Backup Strategy**: Regular backups of persistent data
4. **Access Modes**: Choose appropriate access modes
5. **Monitoring**: Monitor storage usage and performance
`,
	},
}

// InitStorageDocumentation initializes storage documentation resources
func InitStorageDocumentation(mcpServer *server.MCPServer) {
	for _, doc := range storageDocumentationResources {
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
