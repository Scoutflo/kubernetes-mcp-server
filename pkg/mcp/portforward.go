package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
)

// PortForwardState tracks details about an active port forward
type PortForwardState struct {
	Namespace    string
	ResourceName string
	Kind         string
	LocalPorts   []string
	RemotePorts  []string
	StopChan     chan struct{}
	StartTime    time.Time
}

// activePortForwards keeps track of active port forwarding sessions
var (
	activePortForwards = make(map[string]*PortForwardState)
	portForwardMutex   = &sync.Mutex{}
)

func (s *Server) initPortForward() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("create_port_forward",
			mcp.WithDescription("Forward ports from a Kubernetes resource (currently only pods supported) to the local machine. Only available in STDIO mode."),
			mcp.WithString("namespace", mcp.Description("Namespace where the resource is located")),
			mcp.WithString("resource_name", mcp.Description("Name of the resource to port forward"), mcp.Required()),
			mcp.WithString("api_version", mcp.Description("API version of the resource (e.g., 'v1' for pods)")),
			mcp.WithString("kind", mcp.Description("Kind of the resource (e.g., 'Pod')"), mcp.Required()),
			mcp.WithString("ports", mcp.Description("Port mapping in format 'localPort[:remotePort]'. If remotePort is not specified, it will use the same as localPort. Multiple port pairs can be separated by commas.")),
		), Handler: s.portForwardCreate},
		{Tool: mcp.NewTool("cancel_port_forward",
			mcp.WithDescription("Cancel an active port forwarding session. Only available in STDIO mode."),
			mcp.WithString("namespace", mcp.Description("Namespace where the resource is located")),
			mcp.WithString("resource_name", mcp.Description("Name of the resource to stop port forwarding"), mcp.Required()),
		), Handler: s.portForwardCancel},
		{Tool: mcp.NewTool("list_port_forward",
			mcp.WithDescription("List all active port forwarding sessions. Only available in STDIO mode."),
		), Handler: s.portForwardList},
	}
}

func getPortForwardKey(namespace, resourceName string) string {
	return fmt.Sprintf("%s/%s", namespace, resourceName)
}

// checkSTDIOMode returns an error if not running in STDIO mode
func (s *Server) checkSTDIOMode(ctx context.Context) error {
	if !s.IsStdioMode() {
		return errors.New("port forwarding is only available in STDIO mode")
	}
	return nil
}

func (s *Server) portForwardCreate(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if running in STDIO mode
	if err := s.checkSTDIOMode(ctx); err != nil {
		return NewTextResult("", err), nil
	}

	// Extract parameters
	namespace := ctr.Params.Arguments["namespace"]
	if namespace == nil {
		namespace = ""
	}

	resourceName := ctr.Params.Arguments["resource_name"]
	if resourceName == nil {
		return NewTextResult("", errors.New("missing required parameter: resource_name")), nil
	}

	apiVersion := ctr.Params.Arguments["api_version"]
	kind := ctr.Params.Arguments["kind"]
	if kind == nil {
		return NewTextResult("", errors.New("missing required parameter: kind")), nil
	}

	// If api_version is not provided or is incomplete, fetch resource details
	if apiVersion == nil || apiVersion == "" {
		// Default to "v1" for Pods
		if strings.ToLower(kind.(string)) == "pod" {
			apiVersion = "v1"
		} else {
			// Try to get the resource to determine API version
			resourceGetter := server.ServerTool{
				Tool:    mcp.NewTool("resources_get"),
				Handler: s.resourcesGet,
			}

			// Create a temporary request to get the resource
			getRequest := mcp.CallToolRequest{}
			getRequest.Params.Name = "resources_get"
			getRequest.Params.Arguments = map[string]interface{}{
				"namespace": namespace,
				"name":      resourceName,
				"kind":      kind,
				// Try with a common API version for the kind
				"apiVersion": guessAPIVersionForKind(kind.(string)),
			}

			// Call the resource getter
			result, err := resourceGetter.Handler(ctx, getRequest)
			if err != nil || result.IsError {
				return NewTextResult("", fmt.Errorf("failed to auto-detect API version: %v", err)), nil
			}

			// Successfully retrieved the resource, use its API version
			// Note: In a production environment, you'd need to parse the resource YAML
			// to extract the apiVersion properly
			apiVersion = getRequest.Params.Arguments["apiVersion"]
		}
	}

	portsArg := ctr.Params.Arguments["ports"]
	var portMappings []string

	if portsArg == nil || portsArg == "" {
		// Try to auto-detect container ports for the pod
		portMappings = []string{"8080:80"} // Default fallback if nothing can be detected
	} else {
		// Parse ports
		portMappings = strings.Split(portsArg.(string), ",")
		for i, portMapping := range portMappings {
			portMappings[i] = strings.TrimSpace(portMapping)
			parts := strings.Split(portMappings[i], ":")

			// If only local port is provided, use the same for remote port
			if len(parts) == 1 {
				// Make sure it's a valid port
				localPort, err := strconv.Atoi(parts[0])
				if err != nil {
					return NewTextResult("", fmt.Errorf("invalid local port: %s", parts[0])), nil
				}
				// Use the same port for local and remote
				portMappings[i] = fmt.Sprintf("%d:%d", localPort, localPort)
			} else if len(parts) == 2 {
				// Validate both local and remote ports are integers
				_, err := strconv.Atoi(parts[0])
				if err != nil {
					return NewTextResult("", fmt.Errorf("invalid local port: %s", parts[0])), nil
				}
				_, err = strconv.Atoi(parts[1])
				if err != nil {
					return NewTextResult("", fmt.Errorf("invalid remote port: %s", parts[1])), nil
				}
			} else {
				return NewTextResult("", fmt.Errorf("invalid port mapping format: %s. Expected 'localPort[:remotePort]'", portMapping)), nil
			}
		}
	}

	// Check if there's already an active port forward for this resource
	portForwardMutex.Lock()
	portForwardKey := getPortForwardKey(namespace.(string), resourceName.(string))
	if existingState, exists := activePortForwards[portForwardKey]; exists {
		// Close existing port forward
		close(existingState.StopChan)
		delete(activePortForwards, portForwardKey)
	}

	// Initialize channels for port forwarding
	stopChan := make(chan struct{})
	readyChan := make(chan struct{}, 1)

	// Extract local and remote ports for tracking
	var localPorts []string
	var remotePorts []string
	for _, mapping := range portMappings {
		parts := strings.Split(mapping, ":")
		localPorts = append(localPorts, parts[0])
		if len(parts) > 1 {
			remotePorts = append(remotePorts, parts[1])
		} else {
			remotePorts = append(remotePorts, parts[0])
		}
	}

	// Create port forward state
	portForwardState := &PortForwardState{
		Namespace:    namespace.(string),
		ResourceName: resourceName.(string),
		Kind:         kind.(string),
		LocalPorts:   localPorts,
		RemotePorts:  remotePorts,
		StopChan:     stopChan,
		StartTime:    time.Now(),
	}

	// Store the port forward state
	activePortForwards[portForwardKey] = portForwardState
	portForwardMutex.Unlock()

	// Set up output streams
	var stdout strings.Builder

	// Setup port forwarding options
	options := kubernetes.PortForwardOptions{
		Namespace:    namespace.(string),
		ResourceName: resourceName.(string),
		APIVersion:   apiVersion.(string),
		Kind:         kind.(string),
		Ports:        portMappings,
		ReadyChan:    readyChan,
		StopChan:     stopChan,
		Out:          io.MultiWriter(&stdout, os.Stdout),
		ErrOut:       os.Stderr,
	}

	// Start port forwarding in a goroutine
	errChan := make(chan error, 1)
	go func() {
		err := s.k.PortForward(ctx, options)
		if err != nil {
			errChan <- err
		}

		// Clean up when port forwarding ends
		portForwardMutex.Lock()
		delete(activePortForwards, portForwardKey)
		portForwardMutex.Unlock()
	}()

	// Wait for either ready signal or error
	select {
	case <-readyChan:
		return NewTextResult(
			fmt.Sprintf("Port forwarding started successfully for %s/%s\nForwarded ports: %s",
				namespace.(string),
				resourceName.(string),
				strings.Join(portMappings, ", ")),
			nil), nil
	case err := <-errChan:
		portForwardMutex.Lock()
		delete(activePortForwards, portForwardKey)
		portForwardMutex.Unlock()
		return NewTextResult("", fmt.Errorf("port forwarding failed: %v", err)), nil
	case <-ctx.Done():
		portForwardMutex.Lock()
		delete(activePortForwards, portForwardKey)
		portForwardMutex.Unlock()
		return NewTextResult("", errors.New("port forwarding canceled")), nil
	}
}

func (s *Server) portForwardCancel(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if running in STDIO mode
	if err := s.checkSTDIOMode(ctx); err != nil {
		return NewTextResult("", err), nil
	}

	// Extract parameters
	namespace := ctr.Params.Arguments["namespace"]
	if namespace == nil {
		namespace = ""
	}

	resourceName := ctr.Params.Arguments["resource_name"]
	if resourceName == nil {
		return NewTextResult("", errors.New("missing required parameter: resource_name")), nil
	}

	// Check if there's an active port forward for this resource
	portForwardMutex.Lock()
	defer portForwardMutex.Unlock()

	portForwardKey := getPortForwardKey(namespace.(string), resourceName.(string))
	state, exists := activePortForwards[portForwardKey]
	if !exists {
		return NewTextResult("", fmt.Errorf("no active port forwarding found for %s", portForwardKey)), nil
	}

	// Cancel the port forwarding
	close(state.StopChan)
	delete(activePortForwards, portForwardKey)

	return NewTextResult(
		fmt.Sprintf("Port forwarding canceled for %s", portForwardKey),
		nil), nil
}

func (s *Server) portForwardList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if running in STDIO mode
	if err := s.checkSTDIOMode(ctx); err != nil {
		return NewTextResult("", err), nil
	}

	portForwardMutex.Lock()
	defer portForwardMutex.Unlock()

	if len(activePortForwards) == 0 {
		return NewTextResult("No active port forwarding sessions found.", nil), nil
	}

	var sb strings.Builder
	sb.WriteString("Active port forwarding sessions:\n\n")
	sb.WriteString(fmt.Sprintf("%-20s %-20s %-10s %-15s %-15s %-15s\n",
		"NAMESPACE", "RESOURCE", "KIND", "LOCAL PORTS", "REMOTE PORTS", "AGE"))
	sb.WriteString(fmt.Sprintf("%-20s %-20s %-10s %-15s %-15s %-15s\n",
		"----------", "--------", "----", "-----------", "------------", "---"))

	for _, state := range activePortForwards {
		age := time.Since(state.StartTime).Round(time.Second)
		sb.WriteString(fmt.Sprintf("%-20s %-20s %-10s %-15s %-15s %-15s\n",
			state.Namespace,
			state.ResourceName,
			state.Kind,
			strings.Join(state.LocalPorts, ","),
			strings.Join(state.RemotePorts, ","),
			formatDuration(age)))
	}

	return NewTextResult(sb.String(), nil), nil
}

// Helper function to guess the API version based on the kind
func guessAPIVersionForKind(kind string) string {
	kindLower := strings.ToLower(kind)
	switch kindLower {
	case "pod", "service", "namespace", "secret", "configmap", "persistentvolumeclaim", "persistentvolume":
		return "v1"
	case "deployment", "replicaset", "statefulset", "daemonset":
		return "apps/v1"
	case "ingress":
		return "networking.k8s.io/v1"
	default:
		return "v1" // Default fallback
	}
}

// formatDuration formats duration in a human-readable form
func formatDuration(d time.Duration) string {
	if d.Hours() > 24 {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	} else if d.Hours() >= 1 {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
