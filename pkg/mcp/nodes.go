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
		), Handler: s.nodesList},
		{Tool: mcp.NewTool("nodes_get",
			mcp.WithDescription("Get detailed information about a specific Kubernetes node"),
			mcp.WithString("name", mcp.Description("Name of the node"), mcp.Required()),
		), Handler: s.nodesGet},
	}
}

// nodesList handles the nodes_list tool request
func (s *Server) nodesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	klog.V(1).Infof("Tool call: nodes_list - listing all nodes")

	ret, err := s.k.NodesList(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: nodes_list failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list nodes: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: nodes_list completed successfully in %v", duration)
	return NewTextResult(ret, nil), nil
}

// nodesGet handles the nodes_get tool request
func (s *Server) nodesGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	name := ctr.GetString("name", "")
	klog.V(1).Infof("Tool call: nodes_get - name: %s", name)

	if name == "" {
		klog.Errorf("Tool call: nodes_get failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("missing required parameter: name")), nil
	}

	ret, err := s.k.NodesGet(ctx, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: nodes_get failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get node '%s': %v", name, err)), nil
	}

	klog.V(1).Infof("Tool call: nodes_get completed successfully in %v", duration)
	return NewTextResult(ret, nil), nil
}
