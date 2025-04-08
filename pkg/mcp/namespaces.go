package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initNamespaces() []server.ServerTool {
	ret := make([]server.ServerTool, 0)
	ret = append(ret, server.ServerTool{
		Tool: mcp.NewTool("namespaces_list",
			mcp.WithDescription("List all the Kubernetes namespaces in the current cluster"),
		), Handler: s.namespacesList,
	})
	return ret
}

func (s *Server) namespacesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ret, err := s.k.NamespacesList(ctx)
	if err != nil {
		err = fmt.Errorf("failed to list namespaces: %v", err)
	}
	return NewTextResult(ret, err), nil
}
