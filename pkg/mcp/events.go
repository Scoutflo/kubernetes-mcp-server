package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
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
	start := time.Now()
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

	klog.V(1).Infof("Tool: events_list - namespace: %s, involved_object_name: %s, involved_object_kind: %s, involved_object_api_version: %s, field_selectors_count: %d -- got called",
		namespace, involvedObjectName, involvedObjectKind, involvedObjectAPIVersion, len(fieldSelectors))

	ret, err := s.k.EventsList(ctx, namespace, fieldSelectors)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: events_list failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list events: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: events_list completed successfully in %v, result_length: %d", duration, len(ret))
	return NewTextResult(ret, err), nil
}
