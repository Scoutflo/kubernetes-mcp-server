package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initMetricsServer() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("nodes_metrics",
			mcp.WithDescription("Get CPU and memory metrics for all nodes or a specific node"),
			mcp.WithString("name", mcp.Description("Name of the node (optional, if not provided will return metrics for all nodes)")),
		), Handler: s.nodesMetrics},
		{Tool: mcp.NewTool("pods_metrics",
			mcp.WithDescription("Get CPU and memory metrics for pods in a namespace"),
			mcp.WithString("namespace", mcp.Description("Namespace to get pod metrics from (optional, if not provided will use default namespace)")),
			mcp.WithString("name", mcp.Description("Name of the pod (optional, if not provided will return metrics for all pods in the namespace)")),
		), Handler: s.podsMetrics},
	}
}

// nodesMetrics handles the nodes_metrics tool request
func (s *Server) nodesMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	nodeName := ctr.GetString("name", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: nodes_metrics - name=%s - got called by session id: %s", nodeName, sessionID)

	ret, err := s.k.GetNodeMetrics(ctx, nodeName)
	if err != nil {
		duration := time.Since(start)
		if nodeName != "" {
			klog.Errorf("Tool call: nodes_metrics failed after %v: failed to get metrics for node '%s': %v by session id: %s", duration, nodeName, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get metrics for node '%s': %v", nodeName, err)), nil
		}
		klog.Errorf("Tool call: nodes_metrics failed after %v: failed to list node metrics: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list node metrics: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: nodes_metrics completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// podsMetrics handles the pods_metrics tool request
func (s *Server) podsMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	namespace := ctr.GetString("namespace", "")
	podName := ctr.GetString("name", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: pods_metrics - namespace=%s, name=%s - got called by session id: %s", namespace, podName, sessionID)

	ret, err := s.k.GetPodMetrics(ctx, namespace, podName)
	if err != nil {
		duration := time.Since(start)
		if podName != "" {
			klog.Errorf("Tool call: pods_metrics failed after %v: failed to get metrics for pod '%s' in namespace '%s': %v by session id: %s", duration, podName, namespace, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get metrics for pod '%s' in namespace '%s': %v", podName, namespace, err)), nil
		}
		klog.Errorf("Tool call: pods_metrics failed after %v: failed to get pod metrics in namespace '%s': %v by session id: %s", duration, namespace, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get pod metrics in namespace '%s': %v", namespace, err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: pods_metrics completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}
