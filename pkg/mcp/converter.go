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

func (s *Server) initConverters() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("docker_compose_to_k8s_manifest",
			mcp.WithDescription("Converts a Docker Compose file to a Kubernetes manifest. Transforms services, volumes, networks, and other configurations from Docker Compose format to equivalent Kubernetes resources such as Deployments, StatefulSets, Services, ConfigMaps, and more."),
			mcp.WithString("docker_compose", mcp.Description("Docker Compose YAML content to convert to Kubernetes manifest"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Optional namespace to use for all generated Kubernetes resources")),
		), Handler: s.dockerComposeToK8sManifest},
		{Tool: mcp.NewTool("k8s_manifest_to_helm_chart",
			mcp.WithDescription("Converts a Kubernetes manifest to a Helm chart. Transforms Kubernetes resources into templated Helm chart files with parameterized values, following best practices."),
			mcp.WithString("k8s_manifest", mcp.Description("Kubernetes manifest YAML content to convert to a Helm chart"), mcp.Required()),
			mcp.WithString("chart_name", mcp.Description("Optional name for the generated Helm chart")),
		), Handler: s.k8sManifestToHelmChart},
		{Tool: mcp.NewTool("k8s_manifest_to_argo_rollout",
			mcp.WithDescription("Converts a Kubernetes Deployment manifest to an Argo Rollout resource with associated Service resources. Transforms a standard Kubernetes Deployment into an Argo Rollout with advanced deployment strategies like Canary or Blue/Green deployments, and creates all necessary Service resources required for the chosen rollout strategy."),
			mcp.WithString("k8s_manifest", mcp.Description("Kubernetes Deployment manifest YAML content to convert to an Argo Rollout"), mcp.Required()),
			mcp.WithString("strategy", mcp.Description("Rollout strategy to use: 'canary' or 'blueGreen'. For 'blueGreen', both active and preview services will be created. For 'canary', a main service will be created.")),
			mcp.WithString("canary_config", mcp.Description("Optional JSON configuration for canary deployment strategy please use the following if not provided: '{\"steps\":[{\"setWeight\":20},{\"pause\":{\"duration\":\"10\"}},{\"setWeight\":40},{\"pause\":{\"duration\":\"10\"}},{\"setWeight\":60},{\"pause\":{\"duration\":\"10\"}},{\"setWeight\":80},{\"pause\":{\"duration\":\"10\"}}]}'")),
		), Handler: s.k8sManifestToArgoRollout},
	}
}

// dockerComposeToK8sManifest handles the docker_compose_to_k8s_manifest tool request
func (s *Server) dockerComposeToK8sManifest(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract required docker_compose parameter
	sessionID := getSessionID(ctx)
	dockerCompose, err := ctr.RequireString("docker_compose")
	if err != nil {
		klog.Errorf("Tool call: docker_compose_to_k8s_manifest failed after %v: missing required parameter: docker_compose by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: docker_compose")), nil
	}

	// Extract optional namespace parameter
	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: docker_compose_to_k8s_manifest - namespace: %s, docker_compose_length: %d - got called by session id: %s", namespace, len(dockerCompose), sessionID)

	// Generate the Kubernetes manifest
	k8sManifest, err := s.k.DockerComposeToK8sManifest(dockerCompose, namespace)
	duration := time.Since(start)

	if err != nil {
		errMsg := err.Error()
		klog.Errorf("Tool call: docker_compose_to_k8s_manifest failed after %v: %v by session id: %s", duration, err, sessionID)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return NewTextResult("", fmt.Errorf("ERROR: Time limit exceeded while generating the Kubernetes manifest. Please try with a simpler Docker Compose file or try again later.")), nil
		} else if errors.Is(err, context.Canceled) {
			return NewTextResult("", fmt.Errorf("ERROR: The operation was canceled. Please try again.")), nil
		} else if errMsg == "failed to create LLM client" {
			return NewTextResult("", fmt.Errorf("ERROR: Could not connect to LLM service to generate the Kubernetes manifest. The service may be unavailable.")), nil
		}
		return NewTextResult("", fmt.Errorf("ERROR: Failed to convert Docker Compose to Kubernetes manifest: %v", err)), nil
	}

	// Check if the manifest is empty
	if len(k8sManifest) < 50 {
		klog.Errorf("Tool call: docker_compose_to_k8s_manifest failed after %v: generated manifest too short by session id: %s", duration, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: The generated Kubernetes manifest is too short or empty. The Docker Compose file may be invalid or not contain enough information.")), nil
	}

	klog.V(1).Infof("Tool call: docker_compose_to_k8s_manifest completed successfully in %v, manifest_length: %d by session id: %s", duration, len(k8sManifest), sessionID)
	return NewTextResult(k8sManifest, nil), nil
}

// k8sManifestToHelmChart handles the k8s_manifest_to_helm_chart tool request
func (s *Server) k8sManifestToHelmChart(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Extract required k8s_manifest parameter
	k8sManifest, err := ctr.RequireString("k8s_manifest")
	if err != nil {
		klog.Errorf("Tool call: k8s_manifest_to_helm_chart failed after %v: missing required parameter: k8s_manifest by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: k8s_manifest")), nil
	}

	// Extract optional chart_name parameter
	chartName := ctr.GetString("chart_name", "")

	klog.V(1).Infof("Tool: k8s_manifest_to_helm_chart - chart_name: %s, k8s_manifest_length: %d - got called by session id: %s", chartName, len(k8sManifest), sessionID)

	// Generate the Helm chart
	helmChart, err := s.k.K8sManifestToHelmChart(k8sManifest, chartName)
	duration := time.Since(start)

	if err != nil {
		errMsg := err.Error()
		klog.Errorf("Tool call: k8s_manifest_to_helm_chart failed after %v: %v by session id: %s", duration, err, sessionID)
		if errors.Is(err, context.DeadlineExceeded) {
			return NewTextResult("", fmt.Errorf("ERROR: Time limit exceeded while generating the Helm chart. Please try with a simpler Kubernetes manifest or try again later.")), nil
		} else if errors.Is(err, context.Canceled) {
			return NewTextResult("", fmt.Errorf("ERROR: The operation was canceled. Please try again.")), nil
		} else if errMsg == "failed to create LLM client" {
			return NewTextResult("", fmt.Errorf("ERROR: Could not connect to LLM service to generate the Helm chart. The service may be unavailable.")), nil
		}
		return NewTextResult("", fmt.Errorf("ERROR: Failed to convert Kubernetes manifest to Helm chart: %v", err)), nil
	}

	// Check if the chart content is empty
	if len(helmChart) < 50 {
		klog.Errorf("Tool call: k8s_manifest_to_helm_chart failed after %v: generated chart too short by session id: %s", duration, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: The generated Helm chart is too short or empty. The Kubernetes manifest may be invalid or not contain enough information.")), nil
	}

	klog.V(1).Infof("Tool call: k8s_manifest_to_helm_chart completed successfully in %v, chart_length: %d by session id: %s", duration, len(helmChart), sessionID)
	return NewTextResult(helmChart, nil), nil
}

// k8sManifestToArgoRollout handles the k8s_manifest_to_argo_rollout tool request
func (s *Server) k8sManifestToArgoRollout(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Extract required k8s_manifest parameter
	k8sManifest, err := ctr.RequireString("k8s_manifest")
	if err != nil {
		klog.Errorf("Tool call: k8s_manifest_to_argo_rollout failed after %v: missing required parameter: k8s_manifest by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: k8s_manifest")), nil
	}

	// Extract optional strategy parameter
	strategy := ctr.GetString("strategy", "canary") // Default to canary strategy if not specified

	// Extract optional canary_config parameter
	canaryConfig := ctr.GetString("canary_config", "")

	klog.V(1).Infof("Tool call: k8s_manifest_to_argo_rollout - strategy: %s, canary_config_length: %d, k8s_manifest_length: %d - got called by session id: %s", strategy, len(canaryConfig), len(k8sManifest), sessionID)

	// Generate the Argo Rollout manifest
	argoRollout, err := s.k.DeploymentToArgoRollout(k8sManifest, strategy, canaryConfig)
	duration := time.Since(start)

	if err != nil {
		errMsg := err.Error()
		klog.Errorf("Tool call: k8s_manifest_to_argo_rollout failed after %v: %v by session id: %s", duration, err, sessionID)
		if errors.Is(err, context.DeadlineExceeded) {
			return NewTextResult("", fmt.Errorf("ERROR: Time limit exceeded while generating the Argo Rollout. Please try with a simpler Kubernetes manifest or try again later.")), nil
		} else if errors.Is(err, context.Canceled) {
			return NewTextResult("", fmt.Errorf("ERROR: The operation was canceled. Please try again.")), nil
		} else if errMsg == "failed to create LLM client" {
			return NewTextResult("", fmt.Errorf("ERROR: Could not connect to LLM service to generate the Argo Rollout. The service may be unavailable.")), nil
		}
		return NewTextResult("", fmt.Errorf("ERROR: Failed to convert Kubernetes Deployment to Argo Rollout: %v", err)), nil
	}

	// Check if the rollout content is empty
	if len(argoRollout) < 50 {
		klog.Errorf("Tool call: k8s_manifest_to_argo_rollout failed after %v: generated rollout too short by session id: %s", duration, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: The generated Argo Rollout is too short or empty. The Kubernetes manifest may be invalid or not contain enough information.")), nil
	}

	klog.V(1).Infof("Tool call: k8s_manifest_to_argo_rollout completed successfully in %v, rollout_length: %d by session id: %s", duration, len(argoRollout), sessionID)
	return NewTextResult(argoRollout, nil), nil
}
