package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
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
	start := time.Now()
	klog.V(1).Infof("Tool call: namespaces_list - listing all namespaces")

	ret, err := s.k.NamespacesList(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: namespaces_list failed after %v: %v", duration, err)
		err = fmt.Errorf("failed to list namespaces: %v", err)
	} else {
		klog.V(1).Infof("Tool call: namespaces_list completed successfully in %v", duration)
	}

	return NewTextResult(ret, err), nil
}
