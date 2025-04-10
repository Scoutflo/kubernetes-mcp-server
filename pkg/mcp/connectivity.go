package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initConnectivity() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("check_service_connectivity",
			mcp.WithDescription("Check connectivity to a Kubernetes service"),
			mcp.WithString("service_name",
				mcp.Description("Fully qualified service name with port number (e.g. my-service.my-namespace.svc.cluster.local:80)"),
				mcp.Required(),
			),
		), Handler: s.checkServiceConnectivity},
	}
}

func (s *Server) checkServiceConnectivity(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceName, ok := ctr.Params.Arguments["service_name"].(string)
	if !ok || serviceName == "" {
		return NewTextResult("", errors.New("failed to check service connectivity, missing or invalid service_name")), nil
	}

	result, err := s.k.CheckServiceConnectivity(ctx, serviceName)
	if err != nil {
		return NewTextResult("", fmt.Errorf("connectivity check failed: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}
