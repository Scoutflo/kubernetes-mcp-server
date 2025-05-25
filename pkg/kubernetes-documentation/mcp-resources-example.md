# MCP Resources vs Tools - How They Work

This document explains the difference between MCP Tools and MCP Resources, and provides examples of how they work.

## MCP Tools vs Resources

### MCP Tools
- **Purpose**: Actions that can be performed (like functions)
- **Examples**: kubectl commands, cluster operations, deployments
- **Interaction**: Client calls tools with parameters and gets results
- **Nature**: Dynamic, interactive, can modify state
- **Usage**: Similar to REST API endpoints or CLI commands

### MCP Resources
- **Purpose**: Static or dynamic content that can be read (like files or documents)
- **Examples**: Documentation, configuration files, logs, status information
- **Interaction**: Client lists and reads resources by URI
- **Nature**: Generally read-only, informational
- **Usage**: Similar to files in a filesystem or web pages

## Example: How MCP Resources Work

### 1. Resource Definition
In your MCP server, you define resources with:

```go
// Define a resource
resource := mcp.Resource{
    URI:         "k8s://docs/concepts/workloads/pods",
    Name:        "Kubernetes Pods Concepts", 
    Description: "Understanding Kubernetes Pods - the smallest deployable units",
    MIMEType:    "text/markdown",
}

// Create a handler that returns content when the resource is read
handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
    return []mcp.ResourceContents{
        mcp.TextResourceContents{
            URI:      request.Params.URI,
            MIMEType: "text/markdown", 
            Text:     "# Pod Documentation Content...",
        },
    }, nil
}

// Register the resource with the MCP server
server.AddResource(resource, handler)
```

### 2. Client Interaction Flow

#### Step 1: List Available Resources
```bash
# Client requests list of available resources
GET /resources

# Server responds with:
{
  "resources": [
    {
      "uri": "k8s://docs/concepts/workloads/pods",
      "name": "Kubernetes Pods Concepts",
      "description": "Understanding Kubernetes Pods - the smallest deployable units",
      "mimeType": "text/markdown"
    },
    {
      "uri": "k8s://cluster/current", 
      "name": "Current Cluster Information",
      "description": "Information about your current Kubernetes cluster",
      "mimeType": "text/markdown"
    }
  ]
}
```

#### Step 2: Read Specific Resource
```bash
# Client requests to read a specific resource
POST /resources/read
{
  "uri": "k8s://docs/concepts/workloads/pods"
}

# Server responds with content:
{
  "contents": [
    {
      "uri": "k8s://docs/concepts/workloads/pods",
      "mimeType": "text/markdown",
      "text": "# Kubernetes Pods\n\nPods are the smallest deployable units..."
    }
  ]
}
```

## Resource Types in Our Implementation

### 1. Static Documentation Resources
- **URI Pattern**: `k8s://docs/concepts/...`
- **Content**: Pre-written Kubernetes documentation
- **Examples**: Pod concepts, Service networking, ConfigMaps
- **Source**: Defined in `pkg/kubernetes-documentation/` files

### 2. Dynamic Cluster Resources  
- **URI Pattern**: `k8s://cluster/...`
- **Content**: Real-time cluster information
- **Examples**: Current cluster info, available API resources
- **Source**: Generated from live Kubernetes API calls

### 3. Reference Resources
- **URI Pattern**: `k8s://docs/reference/...` 
- **Content**: Command references and guides
- **Examples**: kubectl command reference
- **Source**: Curated reference material

## Structural Organization

Our resource system follows the same organizational pattern as tools:

```go
// In pkg/mcp/documentation.go
func (s *Server) initDocumentationResources() {
    s.initPodDocumentation()        // Pod-related docs
    s.initWorkloadsDocumentation()  // Deployments, Services
    s.initConfigurationDocumentation() // ConfigMaps, Secrets  
    s.initClusterDocumentation()    // Dynamic cluster info
    s.initKubectlDocumentation()    // kubectl commands
}

// Each init function delegates to specialized packages
func (s *Server) initPodDocumentation() {
    kubernetesdocumentation.InitPodDocumentation(s.server)
}
```

This structure mirrors the tools organization:
```
pkg/mcp/
├── pods.go           # Pod tools
├── deployments.go    # Deployment tools  
├── configuration.go  # Configuration tools
└── documentation.go  # Resource initialization

pkg/kubernetes-documentation/
├── pod.go           # Pod documentation resources
├── workloads.go     # Workloads documentation resources
└── configmap.go     # Configuration documentation resources
```

## Why We Added GetClusterInfo() and GetAvailableAPIResourcesList()

These methods were added to support **dynamic resources** that provide real-time cluster information:

1. **GetClusterInfo()**: 
   - Returns current cluster version, context, nodes, namespaces
   - Used by `k8s://cluster/current` resource
   - Provides contextual information about the user's cluster

2. **GetAvailableAPIResourcesList()**:
   - Returns formatted list of available Kubernetes API resources
   - Helps users understand what resources are available in their cluster
   - Different clusters may have different CRDs and API extensions

These enable the MCP client to access not just static documentation, but also live information about the connected Kubernetes cluster.

## Usage Example with MCP Client

When using an MCP client (like Claude Desktop with MCP support):

1. **List resources**: Client sees all available documentation
2. **Read pod docs**: `k8s://docs/concepts/workloads/pods` 
3. **Check cluster**: `k8s://cluster/current` shows live cluster info
4. **Reference commands**: `k8s://docs/reference/kubectl/commands`

This provides a comprehensive knowledge base that combines static documentation with dynamic cluster information, all accessible through the MCP protocol. 