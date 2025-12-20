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

// All pod tools now require k8surl and k8stoken parameters to be provided in each request.
// This allows different clients to connect to different Kubernetes clusters dynamically.
//
// Example usage:
// {
//   "name": "pods_get",
//   "arguments": {
//     "k8surl": "https://your-k8s-api-server:6443",
//     "k8stoken": "your-auth-token"
//   }
// }

type ListResourceToolOutput struct {
	Data                interface{} `json:"data"`
	ContinueToken       string      `json:"continueToken,omitempty"`
	RemainingItemsCount int64       `json:"remainingItemsCount,omitempty"`
}

func (s *Server) initPods() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("pods_get",
			mcp.WithDescription("Get a Kubernetes Pod in the current or provided namespace with the provided name"),
			mcp.WithNumber("limit",
				mcp.DefaultNumber(5),
				mcp.Description("Count of the resources that needs to be listed, this works in additional parameter called 'continue' which will have the value of continue token of paginated data."),
				mcp.Required(),
			),
			mcp.WithString("continue",
				mcp.Description("The continue token that received in previous call with limited count of resource items, this field works with additional field called 'limit'. "),
			),
			mcp.WithString("namespace", mcp.Description("Namespace to get the Pod from")),
			mcp.WithString("name", mcp.Description("Name of the Pod"), mcp.Required()),
		), Handler: s.podsGet},
		{Tool: mcp.NewTool("pods_delete",
			mcp.WithDescription("Delete a Kubernetes Pod in the current or provided namespace with the provided name"),
			mcp.WithString("namespace", mcp.Description("Namespace to delete the Pod from")),
			mcp.WithString("name", mcp.Description("Name of the Pod to delete"), mcp.Required()),
		), Handler: s.podsDelete},
		{Tool: mcp.NewTool("pods_exec",
			mcp.WithDescription("Execute a command in a Kubernetes Pod in the current or provided namespace with the provided name and command"),
			mcp.WithString("namespace", mcp.Description("Namespace to get the Pod from")),
			mcp.WithString("name", mcp.Description("Name of the Pod to get the logs from"), mcp.Required()),
			mcp.WithArray("command", mcp.Description("Command to execute in the Pod container. "+
				"The first item is the command to be run, and the rest are the arguments to that command. "+
				`Example: ["ls", "-l", "/tmp"]`),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
				mcp.Required(),
			),
		), Handler: s.podsExec},
		{Tool: mcp.NewTool("pods_log",
			mcp.WithDescription("Get the logs of a Kubernetes Pod in the current or provided namespace with the provided name"),
			mcp.WithString("namespace", mcp.Description("Namespace to get the Pod logs from")),
			mcp.WithString("name", mcp.Description("Name of the Pod to get the logs from"), mcp.Required()),
			mcp.WithNumber("tail_lines", mcp.Description("Number of lines to get from the end of the logs (Optional, default is 256)")),
			mcp.WithString("start_time", mcp.Description("Start time for log retrieval in RFC3339 format (e.g., '2024-01-01T00:00:00Z') or Unix timestamp. Required if end_time is provided.")),
			mcp.WithString("end_time", mcp.Description("End time for log retrieval in RFC3339 format (e.g., '2024-01-01T23:59:59Z') or Unix timestamp. Required if start_time is provided.")),
			mcp.WithString("time_window", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') - alternative to start_time/end_time. If provided, retrieves logs from (now - time_window) to now.")),
		), Handler: s.podsLog},
		{Tool: mcp.NewTool("pods_run",
			mcp.WithDescription("Run a Kubernetes Pod in the current or provided namespace with the provided container image and optional name"),
			mcp.WithString("namespace", mcp.Description("Namespace to run the Pod in")),
			mcp.WithString("name", mcp.Description("Name of the Pod (Optional, random name if not provided)")),
			mcp.WithString("image", mcp.Description("Container Image to run in the Pod"), mcp.Required()),
			mcp.WithNumber("port", mcp.Description("TCP/IP port to expose from the Pod container (Optional, no port exposed if not provided)")),
		), Handler: s.podsRun},
	}
}

func (s *Server) podsGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	ns := ctr.GetString("namespace", "")
	name := ctr.GetString("name", "")

	sessionID := getSessionID(ctx)

	klog.V(1).Infof("Tool: pods_get - getting pod: %s in namespace: %s - got called by session id: %s", name, ns, sessionID)

	if name == "" {
		klog.Errorf("Tool call: pods_get failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to get pod, missing argument name")), nil
	}

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_get failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsGet(ctx, ns, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: pods_get failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get pod %s in namespace %s: %v", name, ns, err)), nil
	}

	klog.V(1).Infof("Tool call: pods_get completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}

func (s *Server) podsDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	ns := ctr.GetString("namespace", "")
	name := ctr.GetString("name", "")

	sessionID := getSessionID(ctx)

	klog.V(1).Infof("Tool: pods_delete - deleting pod: %s in namespace: %s - got called by session id: %s", name, ns, sessionID)

	if name == "" {
		klog.Errorf("Tool call: pods_delete failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to delete pod, missing argument name")), nil
	}

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_delete failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsDelete(ctx, ns, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: pods_delete failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to delete pod %s in namespace %s: %v", name, ns, err)), nil
	}

	klog.V(1).Infof("Tool call: pods_delete completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}

func (s *Server) podsExec(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	ns := ctr.GetString("namespace", "")
	name := ctr.GetString("name", "")

	if name == "" {
		klog.Errorf("Tool call: pods_exec failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to exec in pod, missing argument name")), nil
	}

	// Get command array using new API
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: pods_exec failed after %v: failed to get arguments", time.Since(start))
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	commandArg, ok := argsMap["command"]
	if !ok {
		klog.Errorf("Tool call: pods_exec failed after %v: missing command argument", time.Since(start))
		return NewTextResult("", errors.New("failed to exec in pod, missing command argument")), nil
	}

	command := make([]string, 0)
	if cmdArray, ok := commandArg.([]interface{}); ok {
		for _, cmd := range cmdArray {
			if strCmd, ok := cmd.(string); ok {
				command = append(command, strCmd)
			}
		}
	} else {
		klog.Errorf("Tool call: pods_exec failed after %v: invalid command argument", time.Since(start))
		return NewTextResult("", errors.New("failed to exec in pod, invalid command argument")), nil
	}

	sessionID := getSessionID(ctx)

	klog.V(1).Infof("Tool: pods_exec - executing command: %v in pod: %s in namespace: %s - got called by session id: %s", command, name, ns, sessionID)

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_exec failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsExec(ctx, ns, name, "", command)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: pods_exec failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to exec in pod %s in namespace %s: %v", name, ns, err)), nil
	} else if ret == "" {
		ret = fmt.Sprintf("The executed command in pod %s in namespace %s has not produced any output", name, ns)
	}

	klog.V(1).Infof("Tool call: pods_exec completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}

func (s *Server) podsLog(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	ns := ctr.GetString("namespace", "")
	name := ctr.GetString("name", "")
	tailLines := ctr.GetFloat("tail_lines", 256)
	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: pods_log - getting logs of pod: %s in namespace: %s with tail lines: %.0f - got called by session id: %s", name, ns, tailLines, sessionID)

	if name == "" {
		klog.Errorf("Tool call: pods_log failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to get pod log, missing argument name")), nil
	}

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: pods_log failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: pods_log failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: pods_log failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: pods_log failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed
	}

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_log failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsLog(ctx, ns, name, int(tailLines), startTime, endTime)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: pods_log failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get pod %s log in namespace %s: %v", name, ns, err)), nil
	} else if ret == "" {
		ret = fmt.Sprintf("The pod %s in namespace %s has not logged any message yet", name, ns)
	}

	klog.V(1).Infof("Tool call: pods_log completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}

func (s *Server) podsRun(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	ns := ctr.GetString("namespace", "")
	name := ctr.GetString("name", "")
	if name == "" {
		name = ""
	}
	image := ctr.GetString("image", "")
	port := ctr.GetFloat("port", 0)
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: pods_run - running pod: %s in namespace: %s with image: %s and port: %.0f - got called by session id: %s", name, ns, image, port, sessionID)

	if image == "" {
		klog.Errorf("Tool call: pods_run failed after %v: missing image parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to run pod, missing argument image")), nil
	}

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_run failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsRun(ctx, ns, name, image, int32(port))
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: pods_run failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to run pod %s in namespace %s: %v", name, ns, err)), nil
	}

	klog.V(1).Infof("Tool call: pods_run completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}
