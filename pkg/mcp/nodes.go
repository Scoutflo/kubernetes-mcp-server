package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initNodes() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("nodes_list",
			mcp.WithDescription("List all Kubernetes nodes in the current cluster"),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.nodesList},
		{Tool: mcp.NewTool("nodes_get",
			mcp.WithDescription("Get detailed information about a specific Kubernetes node"),
			mcp.WithString("name", mcp.Description("Name of the node"), mcp.Required()),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.nodesGet},
	}
}

// nodesList handles the nodes_list tool request
func (s *Server) nodesList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: nodes_list failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: nodes_list - listing all nodes - got called by session id: %s", sessionID)

	ret, err := k.NodesList(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: nodes_list failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list nodes: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: nodes_list completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// nodesGet handles the nodes_get tool request
func (s *Server) nodesGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: nodes_get failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	name := ctr.GetString("name", "")
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: nodes_get - name: %s - got called by session id: %s", name, sessionID)

	if name == "" {
		klog.Errorf("Tool call: nodes_get failed after %v: missing name parameter by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: name")), nil
	}

	ret, err := k.NodesGet(ctx, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: nodes_get failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get node '%s': %v", name, err)), nil
	}

	klog.V(1).Infof("Tool call: nodes_get completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}
