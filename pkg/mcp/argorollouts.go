package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// initArgoRollouts initializes Argo Rollouts tools
func (s *Server) initArgoRollouts() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("create_argo_rollout_config",
				mcp.WithDescription("Generate a YAML configuration for Argo Rollouts with specified deployment strategy"),
				// Required parameters
				mcp.WithString("name",
					mcp.Description("Name of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace for the rollout"),
					mcp.Required(),
				),
				mcp.WithString("image",
					mcp.Description("Container image to deploy"),
					mcp.Required(),
				),
				// Strategy and selector parameters
				mcp.WithString("strategy",
					mcp.Description("Rollout strategy type (canary or blueGreen)"),
					mcp.Required(),
				),
				mcp.WithString("selector_labels",
					mcp.Description("Labels for pod selector, in format: key1=value1,key2=value2"),
				),
				// Resource parameters
				mcp.WithString("replicas",
					mcp.Description("Number of desired pods (default: 1)"),
				),
				mcp.WithString("cpu_request",
					mcp.Description("CPU request for containers (e.g., 100m)"),
				),
				mcp.WithString("memory_request",
					mcp.Description("Memory request for containers (e.g., 256Mi)"),
				),
				// Common parameters
				mcp.WithString("min_ready_seconds",
					mcp.Description("Minimum seconds a pod should be ready before considered available"),
				),
				mcp.WithString("progress_deadline_seconds",
					mcp.Description("Maximum time in seconds for a rollout to make progress"),
				),
				// Blue-Green parameters
				mcp.WithString("active_service",
					mcp.Description("Name of active service (required for blueGreen strategy)"),
				),
				mcp.WithString("preview_service",
					mcp.Description("Name of preview service (optional for blueGreen strategy)"),
				),
				mcp.WithString("auto_promotion_enabled",
					mcp.Description("Whether to automatically promote the new ReplicaSet to active (true/false)"),
				),
				mcp.WithString("auto_promotion_seconds",
					mcp.Description("Seconds to wait before automatically promoting (if auto promotion enabled)"),
				),
				mcp.WithString("scale_down_delay_seconds",
					mcp.Description("Seconds to wait before scaling down the previous ReplicaSet"),
				),
				// Canary parameters
				mcp.WithString("canary_service",
					mcp.Description("Name of canary service (required for canary strategy with traffic routing)"),
				),
				mcp.WithString("stable_service",
					mcp.Description("Name of stable service (required for canary strategy with traffic routing)"),
				),
				mcp.WithString("max_surge",
					mcp.Description("Maximum number of pods that can be scheduled above the desired number"),
				),
				mcp.WithString("max_unavailable",
					mcp.Description("Maximum number of pods that can be unavailable during the update"),
				),
				mcp.WithString("traffic_routing_provider",
					mcp.Description("Traffic routing provider to use (istio, nginx, alb, smi)"),
				),
				mcp.WithString("steps",
					mcp.Description("Canary steps configuration in format: setWeight=10,pause=30s,setWeight=20,pause=1h"),
				),
				mcp.WithString("analysis_templates",
					mcp.Description("Analysis templates to use during rollout, comma-separated"),
				),
			),
			Handler: s.createArgoRolloutsConfig,
		},
		{
			Tool: mcp.NewTool("promote_argo_rollout",
				mcp.WithDescription("Promote an Argo Rollout to advance it to the next step"),
				mcp.WithString("name",
					mcp.Description("Name of the rollout to promote"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("full",
					mcp.Description("If 'true', fully promote the rollout instead of just advancing by one step (blue-green strategy only)"),
				),
			),
			Handler: s.promoteArgoRollout,
		},
		{
			Tool: mcp.NewTool("abort_argo_rollout",
				mcp.WithDescription("Abort an in-progress Argo Rollout and revert to the stable version"),
				mcp.WithString("name",
					mcp.Description("Name of the rollout to abort"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the rollout"),
					mcp.Required(),
				),
			),
			Handler: s.abortArgoRollout,
		},
		{
			Tool: mcp.NewTool("get_argo_rollout",
				mcp.WithDescription("Get the status of an Argo Rollout"),
				mcp.WithString("name",
					mcp.Description("Name of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("output",
					mcp.Description("Output format (json, yaml, wide)"),
				),
			),
			Handler: s.getArgoRollout,
		},
		{
			Tool: mcp.NewTool("set_argo_rollout_weight",
				mcp.WithDescription("Set the canary weight for an Argo Rollout"),
				mcp.WithString("name",
					mcp.Description("Name of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("weight",
					mcp.Description("Weight percentage (0-100) to set for the canary"),
					mcp.Required(),
				),
			),
			Handler: s.setArgoRolloutWeight,
		},
		{
			Tool: mcp.NewTool("pause_argo_rollout",
				mcp.WithDescription("Pause an Argo Rollout to temporarily halt progression"),
				mcp.WithString("name",
					mcp.Description("Name of the rollout to pause"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the rollout"),
					mcp.Required(),
				),
			),
			Handler: s.pauseArgoRollout,
		},
		{
			Tool: mcp.NewTool("set_argo_rollout_image",
				mcp.WithDescription("Set the image for a container in an Argo Rollouts deployment"),
				mcp.WithString("name",
					mcp.Description("Name of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the rollout"),
					mcp.Required(),
				),
				mcp.WithString("container",
					mcp.Description("Name of the container to update (if not specified, updates the first container)"),
				),
				mcp.WithString("image",
					mcp.Description("New image to set for the container"),
					mcp.Required(),
				),
			),
			Handler: s.setArgoRolloutImage,
		},
	}
}

// createArgoRolloutsConfig generates an Argo Rollout YAML configuration based on provided parameters
func (s *Server) createArgoRolloutsConfig(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout name is required")), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		return NewTextResult("", fmt.Errorf("namespace is required")), nil
	}

	image, err := ctr.RequireString("image")
	if err != nil {
		return NewTextResult("", fmt.Errorf("container image is required")), nil
	}

	strategy, err := ctr.RequireString("strategy")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout strategy is required")), nil
	}

	selectorLabels := ctr.GetString("selector_labels", "")
	if selectorLabels == "" {
		return NewTextResult("", fmt.Errorf("selector labels are required")), nil
	}

	// Extract optional parameters with defaults
	replicas := ctr.GetString("replicas", "1")
	minReadySeconds := ctr.GetString("min_ready_seconds", "0")
	progressDeadlineSeconds := ctr.GetString("progress_deadline_seconds", "600")
	cpuRequest := ctr.GetString("cpu_request", "")
	memoryRequest := ctr.GetString("memory_request", "")
	scaleDownDelaySeconds := ctr.GetString("scale_down_delay_seconds", "30")

	// Collect blue-green strategy options
	blueGreenOptions := map[string]string{
		"active_service":         ctr.GetString("active_service", ""),
		"preview_service":        ctr.GetString("preview_service", ""),
		"auto_promotion_enabled": ctr.GetString("auto_promotion_enabled", "false"),
		"auto_promotion_seconds": ctr.GetString("auto_promotion_seconds", ""),
	}

	// Collect canary strategy options
	canaryOptions := map[string]string{
		"max_surge":                ctr.GetString("max_surge", "1"),
		"max_unavailable":          ctr.GetString("max_unavailable", "0"),
		"canary_service":           ctr.GetString("canary_service", ""),
		"stable_service":           ctr.GetString("stable_service", ""),
		"traffic_routing_provider": ctr.GetString("traffic_routing_provider", ""),
		"steps":                    ctr.GetString("steps", ""),
		"analysis_templates":       ctr.GetString("analysis_templates", ""),
	}

	// Create ArgoRolloutsClient
	argoClient, err := s.k.NewArgoRolloutsClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create Argo Rollouts client: %w", err)), nil
	}

	// Generate YAML using the client
	yamlConfig, err := argoClient.GenerateRolloutYAML(
		name, namespace, image, strategy, selectorLabels,
		replicas, minReadySeconds, progressDeadlineSeconds, cpuRequest, memoryRequest, scaleDownDelaySeconds,
		blueGreenOptions, canaryOptions,
	)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(yamlConfig, nil), nil
}

// promoteArgoRollout promotes an Argo Rollout to advance it to the next step
func (s *Server) promoteArgoRollout(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout name is required")), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		return NewTextResult("", fmt.Errorf("namespace is required")), nil
	}

	// Extract optional parameters
	fullPromote := ctr.GetBool("full", false)

	// Create ArgoRolloutsClient
	argoClient, err := s.k.NewArgoRolloutsClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create Argo Rollouts client: %w", err)), nil
	}

	// Promote the rollout
	result, err := argoClient.PromoteRollout(ctx, name, namespace, fullPromote)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// abortArgoRollout aborts an in-progress Argo Rollout and reverts to the stable version
func (s *Server) abortArgoRollout(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout name is required")), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		return NewTextResult("", fmt.Errorf("namespace is required")), nil
	}

	// Create ArgoRolloutsClient
	argoClient, err := s.k.NewArgoRolloutsClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create Argo Rollouts client: %w", err)), nil
	}

	// Abort the rollout
	result, err := argoClient.AbortRollout(ctx, name, namespace)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// getArgoRollout gets the status of an Argo Rollout
func (s *Server) getArgoRollout(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout name is required")), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		return NewTextResult("", fmt.Errorf("namespace is required")), nil
	}

	output := ctr.GetString("output", "")

	// Create ArgoRolloutsClient
	argoClient, err := s.k.NewArgoRolloutsClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create Argo Rollouts client: %w", err)), nil
	}

	// Get the rollout
	rollout, err := argoClient.GetRollout(ctx, name, namespace)
	if err != nil {
		return NewTextResult("", err), nil
	}

	// Format the output
	result, err := argoClient.FormatRolloutOutput(rollout, output)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// setArgoRolloutWeight sets the weight for a canary rollout
func (s *Server) setArgoRolloutWeight(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout name is required")), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		return NewTextResult("", fmt.Errorf("namespace is required")), nil
	}

	weightStr, err := ctr.RequireString("weight")
	if err != nil {
		return NewTextResult("", fmt.Errorf("weight is required")), nil
	}

	weight, err := strconv.Atoi(weightStr)
	if err != nil {
		return NewTextResult("", fmt.Errorf("invalid weight value '%s': %w", weightStr, err)), nil
	}

	// Create ArgoRolloutsClient
	argoClient, err := s.k.NewArgoRolloutsClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create Argo Rollouts client: %w", err)), nil
	}

	// Set the rollout weight
	result, err := argoClient.SetRolloutWeight(ctx, name, namespace, weight)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// pauseArgoRollout pauses an Argo Rollout to temporarily halt progression
func (s *Server) pauseArgoRollout(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout name is required")), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		return NewTextResult("", fmt.Errorf("namespace is required")), nil
	}

	// Create ArgoRolloutsClient
	argoClient, err := s.k.NewArgoRolloutsClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create Argo Rollouts client: %w", err)), nil
	}

	// Pause the rollout
	result, err := argoClient.PauseRollout(ctx, name, namespace)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// setArgoRolloutImage sets the image for a container in an Argo Rollouts deployment
func (s *Server) setArgoRolloutImage(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		return NewTextResult("", fmt.Errorf("rollout name is required")), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		return NewTextResult("", fmt.Errorf("namespace is required")), nil
	}

	image, err := ctr.RequireString("image")
	if err != nil {
		return NewTextResult("", fmt.Errorf("new image is required")), nil
	}

	// Extract optional parameters
	containerName := ctr.GetString("container", "")

	// Create ArgoRolloutsClient
	argoClient, err := s.k.NewArgoRolloutsClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create Argo Rollouts client: %w", err)), nil
	}

	// Set the rollout image
	result, err := argoClient.SetRolloutImage(ctx, name, namespace, containerName, image)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// Helper function to get string parameters with default values
func getStringParam(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

// Helper function for string to bool conversion
func getBoolParam(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key].(string); ok && val != "" {
		lower := strings.ToLower(val)
		if lower == "true" {
			return true
		} else if lower == "false" {
			return false
		}
	}
	return defaultValue
}

// Helper function for string to int conversion
func getIntParam(args map[string]interface{}, key string, defaultValue int) int {
	if val, ok := args[key].(string); ok && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}
