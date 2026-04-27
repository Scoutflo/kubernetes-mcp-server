package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initEvents() []server.ServerTool {
	return []server.ServerTool{
		{Tool: WithMeta(
			mcp.NewTool("events_list",
				mcp.WithDescription("List Kubernetes events. Primary tool for incident triage when a pod, deployment, or node is behaving unexpectedly — events surface scheduling failures, image pull errors, OOM kills, and controller restarts. Always filter by namespace and involved_object_name to reduce noise. Use time_window or start_time/end_time to scope to the incident window."),
				mcp.WithString("namespace",
					mcp.Description("Optional Namespace to retrieve the events from. If not provided, will list events from all namespaces")),
				mcp.WithString("involved_object_name",
					mcp.Description("Optional filter to show events only for resources with this name")),
				mcp.WithString("involved_object_kind",
					mcp.Description("Optional filter to show events only for resources of this kind (e.g. Pod, Deployment)")),
				mcp.WithString("involved_object_api_version",
					mcp.Description("Optional filter to show events only for resources with this apiVersion")),
			),
			map[string]any{"provider": ProviderKubernetes},
		), Handler: s.eventsList},
	}
}

func (s *Server) eventsList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	namespace := ctr.GetString("namespace", "")
	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: events_list failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
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

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: events_list failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: events_list failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: events_list failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: events_list failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: events_list - namespace: %s, involved_object_name: %s, involved_object_kind: %s, involved_object_api_version: %s, field_selectors_count: %d -- got called by session id: %s",
		namespace, involvedObjectName, involvedObjectKind, involvedObjectAPIVersion, len(fieldSelectors), sessionID)

	ret, err := k.EventsList(ctx, namespace, fieldSelectors, startTime, endTime)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: events_list failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list events: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: events_list completed successfully in %v, result_length: %d by session id: %s", duration, len(ret), sessionID)
	return NewTextResult(ret, err), nil
}
