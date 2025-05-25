# MCP Resources Structure

This document explains the structural organization of MCP resources in the kubernetes-mcp-server project.

## Overview

Following the same organizational pattern as tools, the MCP resources are now structured modularly with separate initialization functions and categorized documentation files.

## Structure

### Main Resource Initialization

```go
// pkg/mcp/documentation.go
func (s *Server) initDocumentationResources() {
    s.initPodDocumentation()           // Pod concepts and lifecycle
    s.initWorkloadsDocumentation()     // Deployments, Services
    s.initConfigurationDocumentation() // ConfigMaps, Secrets
    s.initStorageDocumentation()       // PVs, PVCs, Storage Classes
    s.initKubectlDocumentation()       // kubectl command reference
}
```

### Documentation Categories

```
pkg/kubernetes-documentation/
├── pod.go               # Pod documentation resources
├── workloads.go         # Deployments, Services documentation
├── configmap.go         # ConfigMaps and Secrets documentation
├── storage.go           # Storage and PV documentation
├── kubectl-commands.go  # kubectl command reference
└── mcp-resources-example.md   # How MCP resources work
```

## How to Add New Documentation Categories

### 1. Create New Documentation File

```go
// pkg/kubernetes-documentation/networking.go
package kubernetesdocumentation

import (
    "context"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

type NetworkingDocumentation struct {
    URI         string
    Name        string
    Description string
    MIMEType    string
    Content     string
}

var networkingDocumentationResources = []NetworkingDocumentation{
    {
        URI:         "k8s://docs/concepts/networking/ingress",
        Name:        "Kubernetes Ingress",
        Description: "Understanding Ingress for HTTP/HTTPS routing",
        MIMEType:    "text/markdown",
        Content:     `# Ingress Documentation...`,
    },
    // Add more networking resources...
}

func InitNetworkingDocumentation(mcpServer *server.MCPServer) {
    for _, doc := range networkingDocumentationResources {
        resource := mcp.Resource{
            URI:         doc.URI,
            Name:        doc.Name,
            Description: doc.Description,
            MIMEType:    doc.MIMEType,
        }

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
```

### 2. Add Initialization Function

```go
// pkg/mcp/documentation.go

// Add to initDocumentationResources()
func (s *Server) initDocumentationResources() {
    s.initPodDocumentation()
    s.initWorkloadsDocumentation()
    s.initConfigurationDocumentation()
    s.initStorageDocumentation()
    s.initKubectlDocumentation()
    s.initNetworkingDocumentation()  // <-- Add this
}

// Add the init function
func (s *Server) initNetworkingDocumentation() {
    kubernetesdocumentation.InitNetworkingDocumentation(s.server)
}
```

## Resource URI Patterns

- **Static Documentation**: `k8s://docs/concepts/...`
- **Reference Material**: `k8s://docs/reference/...`

## Resource Types

### 1. Static Documentation Resources
- Pre-written Kubernetes concept documentation
- Examples: Pod concepts, Service networking, ConfigMaps
- Source: Defined in `pkg/kubernetes-documentation/` files

### 2. Reference Resources
- Command references and guides
- Examples: kubectl command reference
- Source: Curated reference material

## Benefits of This Structure

1. **Modular Organization**: Each documentation category is in its own file
2. **Easy Extension**: Add new categories by creating new files and init functions
3. **Consistent Pattern**: Follows the same structure as tools
4. **Maintainable**: Documentation is separated from business logic
5. **Testable**: Each category can be tested independently

## MCP Resources vs Tools

### Tools
- **Purpose**: Actions/operations (like kubectl commands)
- **Interaction**: Call with parameters, get results
- **Example**: `kubectl get pods`, `kubectl create deployment`

### Resources
- **Purpose**: Information/documentation
- **Interaction**: List available resources, read specific resource by URI
- **Example**: Pod documentation, kubectl reference

## Client Usage

When using an MCP client:

1. **List Resources**: See all available documentation
2. **Read Specific Resource**: Get detailed content by URI
3. **Browse Categories**: Explore different types of documentation

Example URIs available:
- `k8s://docs/concepts/workloads/pods`
- `k8s://docs/concepts/workloads/deployments`
- `k8s://docs/concepts/configuration/configmap`
- `k8s://docs/concepts/storage/persistent-volumes`
- `k8s://docs/reference/kubectl/commands`

This provides a comprehensive knowledge base accessible through the MCP protocol with static documentation and reference material. 