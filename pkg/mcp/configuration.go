package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
)

func (s *Server) initConfiguration() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("get_available_API_resources",
			mcp.WithDescription("Get all available and supported API resources in the Kubernetes cluster"),
		), Handler: s.getAvailableAPIResources},
	}
}

func (s *Server) getAvailableAPIResources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ret, err := kubernetes.GetAvailableAPIResources(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get available API resources: %v", err)
	}
	return NewTextResult(ret, err), nil
}
