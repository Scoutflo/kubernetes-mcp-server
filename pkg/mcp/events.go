package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initEvents() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("events_list",
			mcp.WithDescription("List all the Kubernetes events in the current cluster from all namespaces with optional filtering by namespace, resource name, kind, or API version"),
			mcp.WithString("namespace",
				mcp.Description("Optional Namespace to retrieve the events from. If not provided, will list events from all namespaces")),
			mcp.WithString("involved_object_name",
				mcp.Description("Optional filter to show events only for resources with this name")),
			mcp.WithString("involved_object_kind",
				mcp.Description("Optional filter to show events only for resources of this kind (e.g. Pod, Deployment)")),
			mcp.WithString("involved_object_api_version",
				mcp.Description("Optional filter to show events only for resources with this apiVersion")),
		), Handler: s.eventsList},
	}
}

func (s *Server) eventsList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ctr.GetString("namespace", "")

	// Extract field selector parameters
	var fieldSelectors []string

	involvedObjectName := ctr.GetString("involved_object_name", "")
	if involvedObjectName != "" {
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("involvedObject.name=%s", involvedObjectName))
	}

	involvedObjectKind := ctr.GetString("involved_object_kind", "")
	if involvedObjectKind != "" {
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("involvedObject.kind=%s", involvedObjectKind))
	}

	involvedObjectAPIVersion := ctr.GetString("involved_object_api_version", "")
	if involvedObjectAPIVersion != "" {
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("involvedObject.apiVersion=%s", involvedObjectAPIVersion))
	}

	ret, err := s.k.EventsList(ctx, namespace, fieldSelectors)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list events: %v", err)), nil
	}
	return NewTextResult(ret, err), nil
}
