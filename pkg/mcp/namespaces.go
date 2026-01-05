package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initNamespaces() []server.ServerTool {
	ret := make([]server.ServerTool, 0)
	ret = append(ret, server.ServerTool{
		Tool: WithMeta(
			mcp.NewTool("namespaces_list",
				mcp.WithDescription("List all Kubernetes namespaces in the cluster. Returns namespace names, status, labels, annotations, and creation timestamps. Use when you need to discover available namespaces, check namespace existence, or audit cluster organization."),
			),
			map[string]any{"provider": ProviderKubernetes},
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

	limit := ctr.GetInt("limit", 10)

	continueToken := ctr.GetString("continue", "")

	ret, freshContinueToken, remainingCount, err := k.NamespacesList(ctx, int64(limit), continueToken)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: namespaces_list failed after %v: %v by session id: %s", duration, err, sessionID)
		err = fmt.Errorf("failed to list namespaces: %v", err)
	} else {
		klog.V(1).Infof("Tool call: namespaces_list completed successfully in %v by session id: %s", duration, sessionID)
	}

	var data interface{}
	if err := json.Unmarshal(ret, &data); err != nil {
		klog.Errorf("Tool call: namespaces_list failed to unmarshal response after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to unmarshal namespace list: %v", err)), nil
	}

	response := ListResourceToolOutput{
		Data:                data,
		ContinueToken:       freshContinueToken,
		RemainingItemsCount: remainingCount,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		klog.Errorf("Tool call: resources_list failed to marshal result to JSON after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to marshal resource list: %v", err)), nil
	}

	return NewTextResult(string(jsonBytes), err), nil
}
