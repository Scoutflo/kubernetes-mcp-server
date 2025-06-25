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
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.namespacesList,
	})
	return ret
}

func (s *Server) namespacesList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: namespaces_list failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: namespaces_list - listing all namespaces - got called by session id: %s", sessionID)

	ret, err := k.NamespacesList(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: namespaces_list failed after %v: %v by session id: %s", duration, err, sessionID)
		err = fmt.Errorf("failed to list namespaces: %v", err)
	} else {
		klog.V(1).Infof("Tool call: namespaces_list completed successfully in %v by session id: %s", duration, sessionID)
	}

	return NewTextResult(ret, err), nil
}
