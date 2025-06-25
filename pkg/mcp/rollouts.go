package mcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initRollouts() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("rollout",
				mcp.WithDescription("The rollout action to perform on the resource (history, pause, restart, resume, status, undo)"),
				mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
				mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
				mcp.WithString("action", mcp.Description("The action to perform on the resource"), mcp.Required()),
				mcp.WithString("resource_type", mcp.Description("The type of resource to rollout (deployment, daemonset, statefulset)"), mcp.Required()),
				mcp.WithString("resource_name", mcp.Description("The name of the resource to rollout"), mcp.Required()),
				mcp.WithString("namespace", mcp.Description("The namespace of the resource (optional, uses default namespace if not provided)")),
				mcp.WithString("revision", mcp.Description("The revision to rollback to (only used with 'undo' action, defaults to previous revision if not specified)")),
			),
			Handler: s.rollout,
		},
	}
}

// rollout handler for the rollout MCP tool
func (s *Server) rollout(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_list_in_namespace failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract required parameters
	action, err := ctr.RequireString("action")
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: rollout failed after %v: missing required parameter: action", duration)
		return NewTextResult("", errors.New("missing required parameter: action")), nil
	}

	resourceType, err := ctr.RequireString("resource_type")
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: rollout failed after %v: missing required parameter: resource_type", duration)
		return NewTextResult("", errors.New("missing required parameter: resource_type")), nil
	}

	resourceName, err := ctr.RequireString("resource_name")
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: rollout failed after %v: missing required parameter: resource_name", duration)
		return NewTextResult("", errors.New("missing required parameter: resource_name")), nil
	}

	// Extract optional parameters
	namespace := ctr.GetString("namespace", "")

	// Handle revision for undo action
	revision := 0
	revStr := ctr.GetString("revision", "")
	if revStr != "" {
		revision, err = strconv.Atoi(revStr)
		if err != nil {
			duration := time.Since(start)
			klog.Errorf("Tool call: rollout failed after %v: invalid revision: %s", duration, revStr)
			return NewTextResult("", fmt.Errorf("invalid revision: %s", revStr)), nil
		}
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: rollout - action=%s, resource_type=%s, resource_name=%s, namespace=%s, revision=%d -- got called by session id: %s",
		action, resourceType, resourceName, namespace, revision, sessionID)

	// Call the Kubernetes rollout function
	result, err := k.ResourceRollout(ctx, namespace, resourceType, resourceName, action, revision)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: rollout failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("rollout failed: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: rollout completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(result, nil), nil
}
