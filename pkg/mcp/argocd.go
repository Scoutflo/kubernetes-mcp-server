package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
	"k8s.io/klog/v2"
)

// initArgoCD initializes ArgoCD tools
func (s *Server) initArgoCD() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("argocd_list_applications",
				mcp.WithDescription("List all ArgoCD-managed applications with their synchronization status and health for cluster-wide oversight"),
				mcp.WithString("project",
					mcp.Description("Filter applications by project name (optional)"),
				),
				mcp.WithString("name",
					mcp.Description("Filter applications by name (optional)"),
				),
				mcp.WithString("repo",
					mcp.Description("Filter applications by repository URL (optional)"),
				),
				mcp.WithString("refresh",
					mcp.Description("Forces application reconciliation if set to 'hard' or 'normal' (optional)"),
				),
			),
			Handler: s.argocdListApplications,
		},
		{
			Tool: mcp.NewTool("argocd_get_application",
				mcp.WithDescription("Retrieve detailed application configuration including sync state and resource status for deployment inspection"),
				mcp.WithString("name",
					mcp.Description("Name of the application"),
					mcp.Required(),
				),
				mcp.WithString("project",
					mcp.Description("Project of the application"),
				),
				mcp.WithString("refresh",
					mcp.Description("Forces application reconciliation if set to 'hard' or 'normal' (optional)"),
				),
			),
			Handler: s.argocdGetApplication,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_events",
				mcp.WithDescription("Access historical events for specific applications to track configuration changes and operational history"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("start_time", mcp.Description("Start time for event retrieval in RFC3339 format (e.g., '2024-01-01T00:00:00Z') or Unix timestamp. Required if end_time is provided.")),
				mcp.WithString("end_time", mcp.Description("End time for event retrieval in RFC3339 format (e.g., '2024-01-01T23:59:59Z') or Unix timestamp. Required if start_time is provided.")),
				mcp.WithString("time_window", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') - alternative to start_time/end_time. If provided, retrieves events from (now - time_window) to now.")),
			),
			Handler: s.argocdGetApplicationEvents,
		},
		{
			Tool: mcp.NewTool("argocd_sync_application",
				mcp.WithDescription("Trigger manual synchronization between Git definitions and cluster resources to enforce desired states"),
				mcp.WithString("name",
					mcp.Description("Name of the application"),
					mcp.Required(),
				),
				mcp.WithString("revision",
					mcp.Description("Revision to sync to (e.g., a branch, tag, or commit SHA)"),
				),
				mcp.WithString("prune",
					mcp.Description("If 'true', prune resources that are no longer defined in Git (accepted values: 'true', 'false')"),
				),
				mcp.WithString("dry_run",
					mcp.Description("If 'true', preview the sync without making changes (accepted values: 'true', 'false')"),
				),
			),
			Handler: s.argocdSyncApplication,
		},
		{
			Tool: mcp.NewTool("argocd_create_application",
				mcp.WithDescription("Define new applications in ArgoCD to establish GitOps workflows for deployment management. Provide parameters as simple strings - the tool automatically constructs the proper ArgoCD nested structure (metadata, spec.source, spec.destination, spec.syncPolicy). Example: name='my-app', repo_url='https://github.com/user/repo.git', path='k8s/', dest_server='https://kubernetes.default.svc', dest_namespace='production'"),
				mcp.WithString("name",
					mcp.Description("The name of the application (string, required). Example: 'my-application'"),
					mcp.Required(),
				),
				mcp.WithString("project",
					mcp.Description("The ArgoCD project name (string, required). Example: 'default'"),
					mcp.Required(),
				),
				mcp.WithString("repo_url",
					mcp.Description("The Git repository URL (string, required). Full HTTPS or SSH URL. Example: 'https://github.com/user/repo.git' or 'git@github.com:user/repo.git'"),
					mcp.Required(),
				),
				mcp.WithString("path",
					mcp.Description("Path within the repository where Kubernetes manifests are located (string, required). Example: 'k8s/' or '.' for root directory"),
					mcp.Required(),
				),
				mcp.WithString("dest_server",
					mcp.Description("Destination Kubernetes API server URL (string, required). Use 'https://kubernetes.default.svc' for in-cluster, or full URL for external clusters. Example: 'https://kubernetes.default.svc'"),
					mcp.Required(),
				),
				mcp.WithString("dest_namespace",
					mcp.Description("Destination Kubernetes namespace where resources will be deployed (string, required). Example: 'production' or 'default'"),
					mcp.Required(),
				),
				mcp.WithString("revision",
					mcp.Description("Git revision to sync (string, optional). Can be branch name, tag, or commit SHA. Default: 'HEAD'. Example: 'main', 'v1.0.0', or 'abc123def'"),
				),
				mcp.WithString("automated_sync",
					mcp.Description("Enable automated sync policy (string, optional). Accepted values: 'true' or 'false'. Default: 'false'. When 'true', ArgoCD will automatically sync when Git changes are detected"),
				),
				mcp.WithString("prune",
					mcp.Description("Enable auto-prune for resources (string, optional). Accepted values: 'true' or 'false'. Default: 'false'. When 'true', resources removed from Git will be automatically deleted from cluster"),
				),
				mcp.WithString("self_heal",
					mcp.Description("Enable self-healing (string, optional). Accepted values: 'true' or 'false'. Default: 'false'. When 'true', ArgoCD will automatically revert manual changes to match Git state"),
				),
				mcp.WithString("validate",
					mcp.Description("Whether to validate the application before creation (string, optional). Accepted values: 'true' or 'false'. Default: 'true'"),
				),
				mcp.WithString("upsert",
					mcp.Description("Whether to update the application if it already exists (string, optional). Accepted values: 'true' or 'false'. Default: 'false'"),
				),
			),
			Handler: s.argocdCreateApplication,
		},
		{
			Tool: mcp.NewTool("argocd_update_application",
				mcp.WithDescription("Modify application configurations to adjust source repositories, target clusters, or sync policies. Provide only the parameters you want to update as simple strings - the tool automatically constructs the proper ArgoCD nested structure. All parameters except 'name' are optional - only provide the ones you want to change"),
				mcp.WithString("name",
					mcp.Description("The application name to update (string, required). Example: 'my-application'"),
					mcp.Required(),
				),
				mcp.WithString("project",
					mcp.Description("New ArgoCD project name (string, optional). Example: 'production'"),
				),
				mcp.WithString("repo_url",
					mcp.Description("New Git repository URL (string, optional). Full HTTPS or SSH URL. Example: 'https://github.com/user/new-repo.git'"),
				),
				mcp.WithString("path",
					mcp.Description("New path within the repository (string, optional). Example: 'k8s/manifests/'"),
				),
				mcp.WithString("dest_server",
					mcp.Description("New destination Kubernetes API server URL (string, optional). Example: 'https://kubernetes.default.svc'"),
				),
				mcp.WithString("dest_namespace",
					mcp.Description("New destination Kubernetes namespace (string, optional). Example: 'staging'"),
				),
				mcp.WithString("revision",
					mcp.Description("New Git revision to sync (string, optional). Can be branch, tag, or commit SHA. Example: 'develop' or 'v2.0.0'"),
				),
				mcp.WithString("automated_sync",
					mcp.Description("Enable/disable automated sync (string, optional). Accepted values: 'true' or 'false'"),
				),
				mcp.WithString("prune",
					mcp.Description("Enable/disable auto-pruning resources (string, optional). Accepted values: 'true' or 'false'"),
				),
				mcp.WithString("self_heal",
					mcp.Description("Enable/disable self-healing (string, optional). Accepted values: 'true' or 'false'"),
				),
				mcp.WithString("validate",
					mcp.Description("Whether to validate the application (string, optional). Accepted values: 'true' or 'false'. Default: 'true'"),
				),
			),
			Handler: s.argocdUpdateApplication,
		},
		{
			Tool: mcp.NewTool("argocd_delete_application",
				mcp.WithDescription("Remove applications from ArgoCD management while preserving underlying Kubernetes resources"),
				mcp.WithString("name",
					mcp.Description("The name of the application to delete"),
					mcp.Required(),
				),
				mcp.WithString("cascade",
					mcp.Description("Whether to delete application resources as well (accepted values: 'true', 'false', default: 'true')"),
				),
				mcp.WithString("propagation_policy",
					mcp.Description("The propagation policy ('foreground', 'background', or 'orphan')"),
				),
			),
			Handler: s.argocdDeleteApplication,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_resource_tree",
				mcp.WithDescription("Retrieve hierarchical dependency relationships between resources within an application"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
			),
			Handler: s.argocdGetApplicationResourceTree,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_managed_resources",
				mcp.WithDescription("List all Kubernetes resources currently managed under specific ArgoCD applications"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
			),
			Handler: s.argocdGetApplicationManagedResources,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_workload_logs",
				mcp.WithDescription("Access container logs from application workloads to monitor runtime behavior and outputs"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("resource_ref",
					mcp.Description("The resource reference in the format of a map with keys 'name', 'namespace', 'kind' (optional), 'group' (optional), 'version' (optional), and 'container' (optional)"),
					mcp.Required(),
				),
				mcp.WithString("tail",
					mcp.Description("Number of lines to show from the end of the logs (default: '100')"),
				),
				mcp.WithString("follow",
					mcp.Description("Follow logs (accepted values: 'true', 'false', default: 'false')"),
				),
				mcp.WithString("start_time", mcp.Description("Start time for log retrieval in RFC3339 format (e.g., '2024-01-01T00:00:00Z') or Unix timestamp. Required if end_time is provided.")),
				mcp.WithString("end_time", mcp.Description("End time for log retrieval in RFC3339 format (e.g., '2024-01-01T23:59:59Z') or Unix timestamp. Required if start_time is provided.")),
				mcp.WithString("time_window", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') - alternative to start_time/end_time. If provided, retrieves logs from (now - time_window) to now.")),
			),
			Handler: s.argocdGetApplicationWorkloadLogs,
		},
		{
			Tool: mcp.NewTool("argocd_get_resource_events",
				mcp.WithDescription("Retrieve event history for individual Kubernetes resources managed by ArgoCD"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("resource_ref",
					mcp.Description("The resource reference in the format of a map with keys 'name', 'namespace', and optional 'uid'"),
					mcp.Required(),
				),
				mcp.WithString("start_time", mcp.Description("Start time for event retrieval in RFC3339 format (e.g., '2024-01-01T00:00:00Z') or Unix timestamp. Required if end_time is provided.")),
				mcp.WithString("end_time", mcp.Description("End time for event retrieval in RFC3339 format (e.g., '2024-01-01T23:59:59Z') or Unix timestamp. Required if start_time is provided.")),
				mcp.WithString("time_window", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') - alternative to start_time/end_time. If provided, retrieves events from (now - time_window) to now.")),
			),
			Handler: s.argocdGetResourceEvents,
		},
		{
			Tool: mcp.NewTool("argocd_get_resource_actions",
				mcp.WithDescription("Discover available operations for specific resources like restart, rollback, or resource hooks"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("resource_ref",
					mcp.Description("The resource reference in the format of a map with keys 'name', 'namespace', 'kind' (optional), 'group' (optional), and 'version' (optional)"),
					mcp.Required(),
				),
			),
			Handler: s.argocdGetResourceActions,
		},
		{
			Tool: mcp.NewTool("argocd_run_resource_action",
				mcp.WithDescription("Execute resource-specific operations such as pod restarts or job retries within applications"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("resource_ref",
					mcp.Description("The resource reference in the format of a map with keys 'name', 'namespace', 'kind', 'group', and 'version'"),
					mcp.Required(),
				),
				mcp.WithString("action",
					mcp.Description("The name of the action to run"),
					mcp.Required(),
				),
			),
			Handler: s.argocdRunResourceAction,
		},
	}
}

// formatJSON formats the JSON output with indentation and optional filtering
func formatJSON(data interface{}) (string, error) {
	resultBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}
	return string(resultBytes), nil
}

// argocdListApplications lists ArgoCD applications with filtering
func (s *Server) argocdListApplications(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_list_applications failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters from the tool request
	project := ctr.GetString("project", "")
	name := ctr.GetString("name", "")
	repo := ctr.GetString("repo", "")
	refresh := ctr.GetString("refresh", "")

	klog.V(1).Infof("Tool: argocd_list_applications - project: %s, name: %s, repo: %s, refresh: %s - got called",
		project, name, repo, refresh)

	// Call the ListApplications method using the K8s Dashboard API
	result, err := k.ListApplications(ctx, project, name, repo, refresh)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_list_applications failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list ArgoCD applications: %w", err)), nil
	}

	// Return the result directly as it's already in JSON format
	if result == "" || result == "null" || result == "{}" {
		klog.V(1).Infof("Tool call: argocd_list_applications completed successfully in %v - no applications found", duration)
		return NewTextResult(`{"applications": [], "count": 0, "message": "No applications found matching the criteria"}`, nil), nil
	}

	klog.V(1).Infof("Tool call: argocd_list_applications completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdGetApplication gets detailed information about a specific application
func (s *Server) argocdGetApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	refresh := ctr.GetString("refresh", "")

	klog.V(1).Infof("Tool: argocd_get_application - name: %s, refresh: %s - got called", name, refresh)

	// Get application details using the K8s Dashboard API
	result, err := k.GetApplication(ctx, name, refresh)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get application '%s': %w", name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_get_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdSyncApplication syncs an ArgoCD application
func (s *Server) argocdSyncApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_sync_application failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_sync_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	revision := ctr.GetString("revision", "")
	prune := ctr.GetBool("prune", false)
	dryRun := ctr.GetBool("dry_run", false)

	klog.V(1).Infof("Tool: argocd_sync_application - name: %s, revision: %s, prune: %t, dry_run: %t - got called",
		name, revision, prune, dryRun)

	// Sync the application using the K8s Dashboard API
	result, err := k.SyncApplication(ctx, name, revision, prune, dryRun)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_sync_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to sync application '%s': %w", name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_sync_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdCreateApplication creates a new ArgoCD application
func (s *Server) argocdCreateApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	project, err := ctr.RequireString("project")
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: missing project parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("project name is required")), nil
	}

	repoURL, err := ctr.RequireString("repo_url")
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: missing repo_url parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("repository URL is required")), nil
	}

	path, err := ctr.RequireString("path")
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: missing path parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("repository path is required")), nil
	}

	destServer, err := ctr.RequireString("dest_server")
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: missing dest_server parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("destination server is required")), nil
	}

	destNamespace, err := ctr.RequireString("dest_namespace")
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: missing dest_namespace parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("destination namespace is required")), nil
	}

	// Extract optional parameters
	revision := ctr.GetString("revision", "")
	if revision == "" {
		revision = "HEAD" // Default revision
	}

	automatedSync := ctr.GetBool("automated_sync", false)
	prune := ctr.GetBool("prune", false)
	selfHeal := ctr.GetBool("self_heal", false)
	validate := ctr.GetBool("validate", true)
	upsert := ctr.GetBool("upsert", false)

	klog.V(1).Infof("Tool: argocd_create_application - name: %s, project: %s, repo_url: %s, path: %s, dest_server: %s, dest_namespace: %s, revision: %s, automated_sync: %t, prune: %t, self_heal: %t, validate: %t, upsert: %t - got called",
		name, project, repoURL, path, destServer, destNamespace, revision, automatedSync, prune, selfHeal, validate, upsert)

	// Create the application using the K8s Dashboard API
	automatedSyncStr := fmt.Sprintf("%t", automatedSync)
	pruneStr := fmt.Sprintf("%t", prune)
	selfHealStr := fmt.Sprintf("%t", selfHeal)
	validateStr := fmt.Sprintf("%t", validate)
	upsertStr := fmt.Sprintf("%t", upsert)

	result, err := k.CreateApplication(ctx, name, project, repoURL, path, destServer, destNamespace,
		revision, automatedSyncStr, pruneStr, selfHealStr, validateStr, upsertStr)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to create application '%s': %w", name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_create_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdUpdateApplication updates an existing ArgoCD application
func (s *Server) argocdUpdateApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_update_application failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_update_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	// Extract optional parameters
	project := ctr.GetString("project", "")
	repoURL := ctr.GetString("repo_url", "")
	path := ctr.GetString("path", "")
	destServer := ctr.GetString("dest_server", "")
	destNamespace := ctr.GetString("dest_namespace", "")
	revision := ctr.GetString("revision", "")
	validateStr := ctr.GetString("validate", "")

	// Parse and convert boolean parameters that might be optional
	var automatedSync, prune, selfHeal *bool

	if automatedSyncStr := ctr.GetString("automated_sync", ""); automatedSyncStr != "" {
		autoSyncVal := strings.ToLower(automatedSyncStr) == "true"
		automatedSync = &autoSyncVal
	}

	if pruneStr := ctr.GetString("prune", ""); pruneStr != "" {
		pruneVal := strings.ToLower(pruneStr) == "true"
		prune = &pruneVal
	}

	if selfHealStr := ctr.GetString("self_heal", ""); selfHealStr != "" {
		selfHealVal := strings.ToLower(selfHealStr) == "true"
		selfHeal = &selfHealVal
	}

	// Default validate to true if not specified
	validate := validateStr == "" || strings.ToLower(validateStr) == "true"

	klog.V(1).Infof("Tool: argocd_update_application - name: %s, project: %s, repo_url: %s, path: %s, dest_server: %s, dest_namespace: %s, revision: %s, validate: %t - got called",
		name, project, repoURL, path, destServer, destNamespace, revision, validate)

	// Update the application using the K8s Dashboard API
	automatedSyncStr := ""
	if automatedSync != nil {
		automatedSyncStr = fmt.Sprintf("%t", *automatedSync)
	}
	pruneStr := ""
	if prune != nil {
		pruneStr = fmt.Sprintf("%t", *prune)
	}
	selfHealStr := ""
	if selfHeal != nil {
		selfHealStr = fmt.Sprintf("%t", *selfHeal)
	}
	validateStr = fmt.Sprintf("%t", validate)

	var result string
	result, err = k.UpdateApplication(ctx, name, project, repoURL, path, destServer, destNamespace,
		revision, automatedSyncStr, pruneStr, selfHealStr, validateStr)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_update_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to update application '%s': %w", name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_update_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdDeleteApplication deletes an ArgoCD application
func (s *Server) argocdDeleteApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_delete_application failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_delete_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	cascade := ctr.GetBool("cascade", true)
	propagationPolicy := ctr.GetString("propagation_policy", "")

	// Default cascade to true if not specified

	// Validate propagation policy
	if propagationPolicy != "" &&
		propagationPolicy != "foreground" &&
		propagationPolicy != "background" &&
		propagationPolicy != "orphan" {
		klog.Errorf("Tool call: argocd_delete_application failed after %v: invalid propagation policy: %s", time.Since(start), propagationPolicy)
		return NewTextResult("", fmt.Errorf("invalid propagation policy: must be 'foreground', 'background', or 'orphan'")), nil
	}

	klog.V(1).Infof("Tool: argocd_delete_application - name: %s, cascade: %t, propagation_policy: %s - got called",
		name, cascade, propagationPolicy)

	// Log the operation for debugging
	klog.Infof("Deleting ArgoCD application '%s' (cascade=%t, propagationPolicy=%s)",
		name, cascade, propagationPolicy)

	// Delete the application using the K8s Dashboard API
	cascadeStr := fmt.Sprintf("%t", cascade)
	result, err := k.DeleteApplication(ctx, name, cascadeStr, propagationPolicy)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_delete_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to delete application '%s': %w", name, err)), nil
	}

	klog.Info(fmt.Sprintf("Successfully deleted application '%s'", name))
	klog.V(1).Infof("Tool call: argocd_delete_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdGetApplicationResourceTree returns the resource tree for an application
func (s *Server) argocdGetApplicationResourceTree(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_resource_tree failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_resource_tree failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	klog.V(1).Infof("Tool: argocd_get_application_resource_tree - name: %s - got called", name)

	// Get application resource tree using the K8s Dashboard API
	result, err := k.GetApplicationResourceTree(ctx, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_resource_tree failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get resource tree for application '%s': %w", name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_get_application_resource_tree completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdGetApplicationManagedResources returns the managed resources for an application
func (s *Server) argocdGetApplicationManagedResources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_managed_resources failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_managed_resources failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	klog.V(1).Infof("Tool: argocd_get_application_managed_resources - name: %s - got called", name)

	// Get application managed resources using the K8s Dashboard API
	result, err := k.GetApplicationManagedResources(ctx, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_managed_resources failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get managed resources for application '%s': %w", name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_get_application_managed_resources completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdGetApplicationWorkloadLogs returns logs for application workload
func (s *Server) argocdGetApplicationWorkloadLogs(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("application_name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing application_name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	// Extract resource_ref parameter which could be a string or map
	var resourceRef kubernetes.ResourceRef

	args := ctr.GetRawArguments()
	if argsMap, ok := args.(map[string]interface{}); ok {
		switch ref := argsMap["resource_ref"].(type) {
		case map[string]interface{}:
			// It's already a map, extract fields directly
			resourceName, ok := ref["name"].(string)
			if !ok || resourceName == "" {
				klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing resource_ref.name", time.Since(start))
				return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
			}

			resourceNamespace, ok := ref["namespace"].(string)
			if !ok || resourceNamespace == "" {
				klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing resource_ref.namespace", time.Since(start))
				return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
			}

			resourceKind, ok := ref["kind"].(string)
			if !ok || resourceKind == "" {
				klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing resource_ref.kind", time.Since(start))
				return NewTextResult("", fmt.Errorf("resource_ref.kind is required")), nil
			}

			resourceGroup, _ := ref["group"].(string)
			resourceVersion, _ := ref["version"].(string)
			container, _ := ref["container"].(string)

			resourceRef = kubernetes.ResourceRef{
				Group:     resourceGroup,
				Version:   resourceVersion,
				Kind:      resourceKind,
				Namespace: resourceNamespace,
				Name:      resourceName,
				Container: container,
			}
		case string:
			// It's a JSON string, try to parse it
			if err = json.Unmarshal([]byte(ref), &resourceRef); err != nil {
				klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: failed to parse resource_ref JSON: %v", time.Since(start), err)
				return NewTextResult("", fmt.Errorf("failed to parse resource_ref JSON: %w", err)), nil
			}

			// Validate required fields
			if resourceRef.Name == "" {
				klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing resource_ref.name in JSON", time.Since(start))
				return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
			}
			if resourceRef.Namespace == "" {
				klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing resource_ref.namespace in JSON", time.Since(start))
				return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
			}
			if resourceRef.Kind == "" {
				klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing resource_ref.kind in JSON", time.Since(start))
				return NewTextResult("", fmt.Errorf("resource_ref.kind is required")), nil
			}
		default:
			klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: resource_ref parameter is required", time.Since(start))
			return NewTextResult("", fmt.Errorf("resource_ref is required")), nil
		}
	} else {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: failed to get arguments", time.Since(start))
		return NewTextResult("", fmt.Errorf("failed to get arguments")), nil
	}

	// Get tail lines parameter if provided
	tailStr := ctr.GetString("tail", "100")
	if tailStr == "" {
		tailStr = "100" // Default to 100 lines
	}

	// Get follow parameter if provided
	followStr := ctr.GetString("follow", "false")
	follow := strings.ToLower(followStr) == "true"

	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed
	}

	klog.V(1).Infof("Tool: argocd_get_application_workload_logs - application: %s, resource: %s/%s/%s, tail: %s, follow: %t - got called",
		name, resourceRef.Kind, resourceRef.Namespace, resourceRef.Name, tailStr, follow)

	// Get workload logs using the K8s Dashboard API
	followStr = fmt.Sprintf("%t", follow)
	resourceRefJSON, _ := json.Marshal(resourceRef)

	result, err := k.GetApplicationWorkloadLogs(ctx, name, string(resourceRefJSON), followStr, tailStr, startTime, endTime)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get logs for workload '%s' in application '%s': %w",
			resourceRef.Name, name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_get_application_workload_logs completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdGetResourceEvents returns events for a resource managed by an application
func (s *Server) argocdGetResourceEvents(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("application_name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: missing application_name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: failed to get arguments", time.Since(start))
		return NewTextResult("", fmt.Errorf("failed to get arguments")), nil
	}
	resourceRef := argsMap["resource_ref"].(map[string]interface{})

	resourceName, ok := resourceRef["name"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: missing resource_ref.name", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
	}

	resourceNamespace, ok := resourceRef["namespace"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: missing resource_ref.namespace", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
	}

	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: argocd_get_resource_events failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: argocd_get_resource_events failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: argocd_get_resource_events failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: argocd_get_resource_events failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed
	}

	klog.V(1).Infof("Tool: argocd_get_resource_events - application: %s, resource: %s/%s - got called",
		name, resourceNamespace, resourceName)

	// Get resource events using the K8s Dashboard API
	resourceRefJSON, _ := json.Marshal(map[string]interface{}{
		"name":      resourceName,
		"namespace": resourceNamespace,
	})

	result, err := k.GetApplicationResourceEvents(ctx, name, string(resourceRefJSON), startTime, endTime)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get events for resource '%s' in application '%s': %w",
			resourceName, name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_get_resource_events completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// formatTimeAgo formats a duration in a human-readable format for the "time ago" use case
func formatTimeAgo(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// argocdGetResourceActions returns available actions for a resource managed by an application
func (s *Server) argocdGetResourceActions(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	// Extract resource parameters from resourceRef object
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: failed to get arguments", time.Since(start))
		return NewTextResult("", fmt.Errorf("failed to get arguments")), nil
	}
	resourceRef := argsMap["resource_ref"].(map[string]interface{})

	resourceName, ok := resourceRef["name"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.name", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
	}

	resourceNamespace, ok := resourceRef["namespace"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.namespace", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
	}

	resourceKind, ok := resourceRef["kind"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.kind", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.kind is required")), nil
	}

	resourceGroup, ok := resourceRef["group"].(string)
	if !ok {
		resourceGroup = "" // group can be empty for core resources
	}

	resourceVersion, ok := resourceRef["version"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.version", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.version is required")), nil
	}

	klog.V(1).Infof("Tool: argocd_get_resource_actions - application: %s, resource: %s/%s/%s/%s/%s - got called",
		name, resourceGroup, resourceVersion, resourceKind, resourceNamespace, resourceName)

	// Get resource actions using the K8s Dashboard API
	resourceRefJSON, _ := json.Marshal(map[string]interface{}{
		"name":      resourceName,
		"namespace": resourceNamespace,
		"kind":      resourceKind,
		"group":     resourceGroup,
		"version":   resourceVersion,
	})

	result, err := k.GetResourceActions(ctx, name, string(resourceRefJSON))
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get actions for resource '%s' in application '%s': %w",
			resourceName, name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_get_resource_actions completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdRunResourceAction runs an action on a resource managed by an application
func (s *Server) argocdRunResourceAction(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	// Extract resource parameters from resourceRef object
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: failed to get arguments", time.Since(start))
		return NewTextResult("", fmt.Errorf("failed to get arguments")), nil
	}
	resourceRef := argsMap["resource_ref"].(map[string]interface{})

	resourceName, ok := resourceRef["name"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing resource_ref.name", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
	}

	resourceNamespace, ok := resourceRef["namespace"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing resource_ref.namespace", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
	}

	resourceKind, ok := resourceRef["kind"].(string)
	if !ok {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing resource_ref.kind", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.kind is required")), nil
	}

	resourceGroup, ok := resourceRef["group"].(string)
	if !ok {
		resourceGroup = "" // group can be empty for core resources
	}

	resourceVersion, ok := resourceRef["version"].(string)

	action, err := ctr.RequireString("action")
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing action parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("action name is required")), nil
	}

	klog.V(1).Infof("Tool: argocd_run_resource_action - application: %s, resource: %s/%s/%s/%s/%s, action: %s - got called",
		name, resourceGroup, resourceVersion, resourceKind, resourceNamespace, resourceName, action)

	// Run the action using the K8s Dashboard API
	resourceRefJSON, _ := json.Marshal(map[string]interface{}{
		"name":      resourceName,
		"namespace": resourceNamespace,
		"kind":      resourceKind,
		"group":     resourceGroup,
		"version":   resourceVersion,
	})

	result, err := k.RunResourceAction(ctx, name, string(resourceRefJSON), action)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to run action '%s' on resource '%s' in application '%s': %w",
			action, resourceName, name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_run_resource_action completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdGetApplicationEvents returns events for an application
func (s *Server) argocdGetApplicationEvents(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract parameters
	name, err := ctr.RequireString("application_name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed after %v: missing application_name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: argocd_get_application_events failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: argocd_get_application_events failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: argocd_get_application_events failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: argocd_get_application_events failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed
	}

	klog.V(1).Infof("Tool: argocd_get_application_events - application: %s - got called", name)

	// Get application events using the K8s Dashboard API
	result, err := k.GetApplicationEvents(ctx, name, startTime, endTime)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get events for application '%s': %w", name, err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_get_application_events completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}
