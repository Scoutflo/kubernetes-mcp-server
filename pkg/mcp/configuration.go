package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initConfiguration() []server.ServerTool {
	return []server.ServerTool{
		{Tool: WithMeta(
			mcp.NewTool("get_available_API_resources",
				mcp.WithDescription("Retrieve all available API resources supported by the Kubernetes cluster. Returns a list of API resources with their group, version, kind, namespaced status, and supported verbs. Use when you need to discover what resources are available, check API capabilities, verify resource types, or understand cluster API structure. No parameters required."),
			),
			map[string]any{"provider": ProviderKubernetes},
		), Handler: s.getAvailableAPIResources},
	}
}

func (s *Server) getAvailableAPIResources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: get_available_API_resources failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: get_available_API_resources - got called by session id: %s", sessionID)
	ret, err := k.GetAvailableAPIResources(ctx)
	if err != nil {
		klog.Errorf("Tool call: get_available_API_resources failed after %v: %v by session id: %s", time.Since(start), err, sessionID)
		err = fmt.Errorf("failed to get available API resources: %v", err)
	}
	klog.V(1).Infof("Tool call: get_available_API_resources completed successfully in %v by session id: %s", time.Since(start), sessionID)
	return NewTextResult(ret, err), nil
}
