package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initNamespaces() []server.ServerTool {
	ret := make([]server.ServerTool, 0)
	ret = append(ret, server.ServerTool{
		Tool: mcp.NewTool("namespaces_list",
			mcp.WithDescription("List all the Kubernetes namespaces in the current cluster"),
			mcp.WithNumber("limit",
				mcp.DefaultNumber(10),
				mcp.Description("Count of the resources that needs to be listed, this works in additional parameter called 'continue' which will have the value of continue token of paginated data."),
				mcp.Required(),
			),
			mcp.WithNumber("continue",
				mcp.Description("The continue token that received in previous call with limited count of resource items, this field works with additional field called 'limit'. "),
			),
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

	limitStr, limitErr := ctr.RequireString("limit")
	var limit int64 = 0
	if limitErr == nil && limitStr != "" {
		parsedLimit, parseErr := strconv.ParseInt(limitStr, 10, 64)
		if parseErr == nil {
			limit = parsedLimit
		}
	}
	continueToken, continueErr := ctr.RequireString("continue")
	if continueErr != nil {
		continueToken = ""
	}

	ret, freshContinueToken, err := k.NamespacesList(ctx, limit, continueToken)
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

	response := struct {
		Data          interface{} `json:"data"`
		ContinueToken string      `json:"continueToken,omitempty"`
	}{
		Data:          data,
		ContinueToken: freshContinueToken,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		klog.Errorf("Tool call: resources_list failed to marshal result to JSON after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to marshal resource list: %v", err)), nil
	}

	return NewTextResult(string(jsonBytes), err), nil
}
