package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	ret, err := s.k.NodesList(ctx)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list nodes: %v", err)), nil
	}
	return NewTextResult(ret, nil), nil
}

// nodesGet handles the nodes_get tool request
func (s *Server) nodesGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := ctr.GetString("name", "")
	if name == "" {
		return NewTextResult("", errors.New("missing required parameter: name")), nil
	}

	ret, err := s.k.NodesGet(ctx, name)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get node '%s': %v", name, err)), nil
	}
	return NewTextResult(ret, nil), nil
}
