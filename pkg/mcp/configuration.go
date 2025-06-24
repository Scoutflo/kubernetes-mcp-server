package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
	"k8s.io/klog/v2"
)

func (s *Server) initConfiguration() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("get_available_API_resources",
			mcp.WithDescription("Get all available and supported API resources in the Kubernetes cluster"),
		), Handler: s.getAvailableAPIResources},
	}
}

func (s *Server) getAvailableAPIResources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: get_available_API_resources - got called by session id: %s", sessionID)
	ret, err := kubernetes.GetAvailableAPIResources(ctx)
	if err != nil {
		klog.Errorf("Tool call: get_available_API_resources failed after %v: %v by session id: %s", time.Since(start), err, sessionID)
		err = fmt.Errorf("failed to get available API resources: %v", err)
	}
	klog.V(1).Infof("Tool call: get_available_API_resources completed successfully in %v by session id: %s", time.Since(start), sessionID)
	return NewTextResult(ret, err), nil
}
