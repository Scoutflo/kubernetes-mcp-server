package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	nodeName := ctr.GetString("name", "")

	ret, err := s.k.GetNodeMetrics(ctx, nodeName)
	if err != nil {
		if nodeName != "" {
			return NewTextResult("", fmt.Errorf("failed to get metrics for node '%s': %v", nodeName, err)), nil
		}
		return NewTextResult("", fmt.Errorf("failed to list node metrics: %v", err)), nil
	}
	return NewTextResult(ret, nil), nil
}

// podsMetrics handles the pods_metrics tool request
func (s *Server) podsMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ctr.GetString("namespace", "")

	podName := ctr.GetString("name", "")

	ret, err := s.k.GetPodMetrics(ctx, namespace, podName)
	if err != nil {
		if podName != "" {
			return NewTextResult("", fmt.Errorf("failed to get metrics for pod '%s' in namespace '%s': %v", podName, namespace, err)), nil
		}
		return NewTextResult("", fmt.Errorf("failed to get pod metrics in namespace '%s': %v", namespace, err)), nil
	}
	return NewTextResult(ret, nil), nil
}
