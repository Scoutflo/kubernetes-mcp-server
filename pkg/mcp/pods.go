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
//   "name": "pods_list",
//   "arguments": {
//     "k8surl": "https://your-k8s-api-server:6443",
//     "k8stoken": "your-auth-token"
//   }
// }

func (s *Server) initPods() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("pods_list",
			mcp.WithDescription("List all the Kubernetes pods in the current cluster from all namespaces"),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.podsListInAllNamespaces},
		{Tool: mcp.NewTool("pods_list_in_namespace",
			mcp.WithDescription("List all the Kubernetes pods in the specified namespace in the current cluster"),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to list pods from"), mcp.Required()),
		), Handler: s.podsListInNamespace},
		{Tool: mcp.NewTool("pods_get",
			mcp.WithDescription("Get a Kubernetes Pod in the current or provided namespace with the provided name"),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to get the Pod from")),
			mcp.WithString("name", mcp.Description("Name of the Pod"), mcp.Required()),
		), Handler: s.podsGet},
		{Tool: mcp.NewTool("pods_delete",
			mcp.WithDescription("Delete a Kubernetes Pod in the current or provided namespace with the provided name"),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to delete the Pod from")),
			mcp.WithString("name", mcp.Description("Name of the Pod to delete"), mcp.Required()),
		), Handler: s.podsDelete},
		{Tool: mcp.NewTool("pods_exec",
			mcp.WithDescription("Execute a command in a Kubernetes Pod in the current or provided namespace with the provided name and command"),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
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
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to get the Pod logs from")),
			mcp.WithString("name", mcp.Description("Name of the Pod to get the logs from"), mcp.Required()),
			mcp.WithNumber("tail_lines", mcp.Description("Number of lines to get from the end of the logs (Optional, default is 256)")),
		), Handler: s.podsLog},
		{Tool: mcp.NewTool("pods_run",
			mcp.WithDescription("Run a Kubernetes Pod in the current or provided namespace with the provided container image and optional name"),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to run the Pod in")),
			mcp.WithString("name", mcp.Description("Name of the Pod (Optional, random name if not provided)")),
			mcp.WithString("image", mcp.Description("Container Image to run in the Pod"), mcp.Required()),
			mcp.WithNumber("port", mcp.Description("TCP/IP port to expose from the Pod container (Optional, no port exposed if not provided)")),
		), Handler: s.podsRun},
	}
}

func (s *Server) podsListInAllNamespaces(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: pods_list - listing all pods in all namespaces - got called by session id: %s", sessionID)

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_list failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsListInAllNamespaces(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: pods_list failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list pods in all namespaces: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: pods_list completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}

func (s *Server) podsListInNamespace(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	ns := ctr.GetString("namespace", "")

	sessionID := getSessionID(ctx)

	klog.V(1).Infof("Tool: pods_list_in_namespace - listing all pods in namespace: %s - got called by session id: %s", ns, sessionID)

	if ns == "" {
		klog.Errorf("Tool call: pods_list_in_namespace failed after %v: missing namespace parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to list pods in namespace, missing argument namespace")), nil
	}

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_list_in_namespace failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsListInNamespace(ctx, ns)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: pods_list_in_namespace failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list pods in namespace %s: %v", ns, err)), nil
	}

	klog.V(1).Infof("Tool call: pods_list_in_namespace completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
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
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: pods_log - getting logs of pod: %s in namespace: %s with tail lines: %.0f - got called by session id: %s", name, ns, tailLines, sessionID)

	if name == "" {
		klog.Errorf("Tool call: pods_log failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to get pod log, missing argument name")), nil
	}

	// Get Kubernetes client from request parameters
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: pods_log failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	ret, err := k.PodsLog(ctx, ns, name, int(tailLines))
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
