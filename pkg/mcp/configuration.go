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
		{Tool: mcp.NewTool("configuration_view",
			mcp.WithDescription("Get the current Kubernetes configuration content as a kubeconfig YAML"),
			mcp.WithBoolean("minified", mcp.Description("Return a minified version of the configuration. "+
				"If set to true, keeps only the current-context and the relevant pieces of the configuration for that context. "+
				"If set to false, all contexts, clusters, auth-infos, and users are returned in the configuration. "+
				"(Optional, default true)")),
		), Handler: configurationView},
		{Tool: mcp.NewTool("get_available_API_resources",
			mcp.WithDescription("Get all available and supported API resources in the Kubernetes cluster"),
		), Handler: s.getAvailableAPIResources},
	}
}

func configurationView(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	minify := true
	minified := ctr.GetBool("minified", true)
	if minified {
		minify = minified
	}
	ret, err := kubernetes.ConfigurationView(minify)
	if err != nil {
		err = fmt.Errorf("failed to get configuration: %v", err)
	}
	return NewTextResult(ret, err), nil
}

func (s *Server) getAvailableAPIResources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ret, err := kubernetes.GetAvailableAPIResources(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get available API resources: %v", err)
	}
	return NewTextResult(ret, err), nil
}
