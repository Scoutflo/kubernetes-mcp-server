package mcp

import (
	"context"
	"encoding/json"
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
				mcp.WithDescription("List applications in ArgoCD with filtering options"),
				mcp.WithString("project",
					mcp.Description("Filter applications by project name (optional)"),
				),
				mcp.WithString("name",
					mcp.Description("Filter applications by name (optional)"),
				),
				mcp.WithString("repo",
					mcp.Description("Filter applications by repository URL (optional)"),
				),
				mcp.WithString("namespace",
					mcp.Description("Filter applications by namespace"),
					mcp.Required(),
				),
				mcp.WithString("refresh",
					mcp.Description("Forces application reconciliation if set to 'hard' or 'normal' (optional)"),
				),
			),
			Handler: s.argocdListApplications,
		},
		{
			Tool: mcp.NewTool("argocd_get_application",
				mcp.WithDescription("Get detailed information about a specific ArgoCD application"),
				mcp.WithString("name",
					mcp.Description("Name of the application"),
					mcp.Required(),
				),
				mcp.WithString("project",
					mcp.Description("Project of the application"),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the application"),
					mcp.Required(),
				),
				mcp.WithString("refresh",
					mcp.Description("Forces application reconciliation if set to 'hard' or 'normal' (optional)"),
				),
			),
			Handler: s.argocdGetApplication,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_events",
				mcp.WithDescription("Returns events for an application"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("application_namespace",
					mcp.Description("The application namespace (optional)"),
				),
			),
			Handler: s.argocdGetApplicationEvents,
		},
		{
			Tool: mcp.NewTool("argocd_sync_application",
				mcp.WithDescription("Sync an ArgoCD application to its desired state"),
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
				mcp.WithString("namespace",
					mcp.Description("Namespace of the application"),
					mcp.Required(),
				),
				mcp.WithArray("resources",
					mcp.Description("List of specific resources to sync (optional), format: [\"group:kind:name\"]"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
			),
			Handler: s.argocdSyncApplication,
		},
		{
			Tool: mcp.NewTool("argocd_create_application",
				mcp.WithDescription("Create a new application in ArgoCD"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("project",
					mcp.Description("The project name"),
					mcp.Required(),
				),
				mcp.WithString("repo_url",
					mcp.Description("The Git repository URL"),
					mcp.Required(),
				),
				mcp.WithString("path",
					mcp.Description("Path within the repository"),
					mcp.Required(),
				),
				mcp.WithString("dest_server",
					mcp.Description("Destination K8s API server URL"),
					mcp.Required(),
				),
				mcp.WithString("dest_namespace",
					mcp.Description("Destination namespace"),
					mcp.Required(),
				),
				mcp.WithString("revision",
					mcp.Description("Git revision (default: HEAD)"),
				),
				mcp.WithString("automated_sync",
					mcp.Description("Enable automated sync (accepted values: 'true', 'false', default: 'false')"),
				),
				mcp.WithString("prune",
					mcp.Description("Auto-prune resources (accepted values: 'true', 'false', default: 'false')"),
				),
				mcp.WithString("self_heal",
					mcp.Description("Enable self-healing (accepted values: 'true', 'false', default: 'false')"),
				),
				mcp.WithString("namespace",
					mcp.Description("Application namespace"),
				),
				mcp.WithString("validate",
					mcp.Description("Whether to validate the application before creation (accepted values: 'true', 'false', default: 'true')"),
				),
				mcp.WithString("upsert",
					mcp.Description("Whether to update the application if it already exists (accepted values: 'true', 'false', default: 'false')"),
				),
			),
			Handler: s.argocdCreateApplication,
		},
		{
			Tool: mcp.NewTool("argocd_update_application",
				mcp.WithDescription("Update an existing application in ArgoCD"),
				mcp.WithString("name",
					mcp.Description("The application name to update"),
					mcp.Required(),
				),
				mcp.WithString("project",
					mcp.Description("New project name (optional)"),
				),
				mcp.WithString("repo_url",
					mcp.Description("New Git repository URL (optional)"),
				),
				mcp.WithString("path",
					mcp.Description("New path within the repository (optional)"),
				),
				mcp.WithString("dest_server",
					mcp.Description("New destination K8s API server URL (optional)"),
				),
				mcp.WithString("dest_namespace",
					mcp.Description("New destination namespace (optional)"),
				),
				mcp.WithString("revision",
					mcp.Description("New Git revision (optional)"),
				),
				mcp.WithString("automated_sync",
					mcp.Description("Enable/disable automated sync (accepted values: 'true', 'false')"),
				),
				mcp.WithString("prune",
					mcp.Description("Enable/disable auto-pruning resources (accepted values: 'true', 'false')"),
				),
				mcp.WithString("self_heal",
					mcp.Description("Enable/disable self-healing (accepted values: 'true', 'false')"),
				),
				mcp.WithString("validate",
					mcp.Description("Whether to validate the application (accepted values: 'true', 'false', default: 'true')"),
				),
			),
			Handler: s.argocdUpdateApplication,
		},
		{
			Tool: mcp.NewTool("argocd_delete_application",
				mcp.WithDescription("Delete an application from ArgoCD"),
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
				mcp.WithString("namespace",
					mcp.Description("The application namespace"),
					mcp.Required(),
				),
			),
			Handler: s.argocdDeleteApplication,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_resource_tree",
				mcp.WithDescription("Returns resource tree for application by application name"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("The application namespace"),
					mcp.Required(),
				),
			),
			Handler: s.argocdGetApplicationResourceTree,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_managed_resources",
				mcp.WithDescription("Returns managed resources for application by application name"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("The application namespace"),
					mcp.Required(),
				),
			),
			Handler: s.argocdGetApplicationManagedResources,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_workload_logs",
				mcp.WithDescription("Returns logs for application workload (Deployment, StatefulSet, Pod, etc.) by application name and resource details"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("application_namespace",
					mcp.Description("The namespace of the application"),
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
			),
			Handler: s.argocdGetApplicationWorkloadLogs,
		},
		{
			Tool: mcp.NewTool("argocd_get_resource_events",
				mcp.WithDescription("Returns events for a resource that is managed by an application"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("application_namespace",
					mcp.Description("The namespace of the application"),
					mcp.Required(),
				),
				mcp.WithString("resource_ref",
					mcp.Description("The resource reference in the format of a map with keys 'name', 'namespace', and optional 'uid'"),
					mcp.Required(),
				),
			),
			Handler: s.argocdGetResourceEvents,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_events",
				mcp.WithDescription("Returns events for an application"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("application_namespace",
					mcp.Description("The application namespace (optional)"),
				),
			),
			Handler: s.argocdGetApplicationEvents,
		},
		{
			Tool: mcp.NewTool("argocd_get_resource_actions",
				mcp.WithDescription("Returns actions for a resource that is managed by an application"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("app_namespace",
					mcp.Description("The namespace of the application"),
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
				mcp.WithDescription("Runs an action on a resource"),
				mcp.WithString("name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("app_namespace",
					mcp.Description("The namespace of the application"),
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
	// Extract parameters from the tool request
	project := ctr.GetString("project", "")
	name := ctr.GetString("name", "")
	repo := ctr.GetString("repo", "")
	namespace := ctr.GetString("namespace", "")
	refresh := ctr.GetString("refresh", "")

	klog.V(1).Infof("Tool: argocd_list_applications - project: %s, name: %s, repo: %s, namespace: %s, refresh: %s - got called",
		project, name, repo, namespace, refresh)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_list_applications failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Call the ListApplications method with individual parameters
	appList, err := argoClient.ListApplications(ctx, project, name, repo, namespace, refresh)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_list_applications failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list ArgoCD applications: %w", err)), nil
	}

	// Format the output as compact JSON
	if len(appList.Items) == 0 {
		klog.V(1).Infof("Tool call: argocd_list_applications completed successfully in %v - no applications found", duration)
		return NewTextResult(`{"applications": [], "count": 0, "message": "No applications found matching the criteria"}`, nil), nil
	}

	// Create a simplified structure for easier reading
	type ApplicationSummary struct {
		Name           string `json:"name"`
		Project        string `json:"project"`
		SyncStatus     string `json:"sync_status"`
		HealthStatus   string `json:"health_status"`
		RepoURL        string `json:"repo_url"`
		Path           string `json:"path"`
		TargetRevision string `json:"target_revision"`
		Namespace      string `json:"namespace"`
		DestNamespace  string `json:"dest_namespace"`
		Server         string `json:"dest_server"`
	}

	type ApplicationListResponse struct {
		Applications []ApplicationSummary `json:"applications"`
		Count        int                  `json:"count"`
		Filters      map[string]string    `json:"filters,omitempty"`
	}

	var summaries []ApplicationSummary
	for _, app := range appList.Items {
		syncStatus := "Unknown"
		if app.Status.Sync.Status != "" {
			syncStatus = app.Status.Sync.Status
		}

		healthStatus := "Unknown"
		if app.Status.Health.Status != "" {
			healthStatus = app.Status.Health.Status
		}

		summary := ApplicationSummary{
			Name:           app.Metadata.Name,
			Project:        app.Spec.Project,
			SyncStatus:     syncStatus,
			HealthStatus:   healthStatus,
			RepoURL:        app.Spec.Source.RepoURL,
			Path:           app.Spec.Source.Path,
			TargetRevision: app.Spec.Source.TargetRevision,
			Namespace:      app.Metadata.Namespace,
			DestNamespace:  app.Spec.Destination.Namespace,
			Server:         app.Spec.Destination.Server,
		}
		summaries = append(summaries, summary)
	}

	// Build filters map
	filters := make(map[string]string)
	if project != "" {
		filters["project"] = project
	}
	if name != "" {
		filters["name"] = name
	}
	if repo != "" {
		filters["repo"] = repo
	}
	if namespace != "" {
		filters["namespace"] = namespace
	}
	if refresh != "" {
		filters["refresh"] = refresh
	}

	response := ApplicationListResponse{
		Applications: summaries,
		Count:        len(summaries),
	}

	if len(filters) > 0 {
		response.Filters = filters
	}

	// Return as compact JSON
	jsonResult, err := json.Marshal(response)
	if err != nil {
		klog.Errorf("Tool call: argocd_list_applications failed after %v: failed to marshal response: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to marshal applications list: %w", err)), nil
	}

	klog.V(1).Infof("Tool call: argocd_list_applications completed successfully in %v - found %d applications", duration, len(summaries))
	return NewTextResult(string(jsonResult), nil), nil
}

// argocdGetApplication gets detailed information about a specific application
func (s *Server) argocdGetApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace := ctr.GetString("namespace", "")
	refresh := ctr.GetString("refresh", "")

	klog.V(1).Infof("Tool: argocd_get_application - name: %s, namespace: %s, refresh: %s - got called", name, namespace, refresh)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application details
	app, err := argoClient.GetApplication(ctx, name, refresh)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get application '%s': %w", name, err)), nil
	}

	// Build a user-friendly summary of the application
	var sb strings.Builder

	// Provide full JSON for detailed information
	sb.WriteString("\nFull Application Details (JSON):\n")

	// Format as JSON with indentation
	jsonResult, err := formatJSON(app)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	sb.WriteString(jsonResult)

	// Add sync policy information
	if app.Spec.SyncPolicy != nil && app.Spec.SyncPolicy.Automated != nil {
		sb.WriteString("Sync Policy: Automated")
		if app.Spec.SyncPolicy.Automated.Prune {
			sb.WriteString(" (with pruning)")
		}
		if app.Spec.SyncPolicy.Automated.SelfHeal {
			sb.WriteString(" (with self-healing)")
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("Sync Policy: Manual\n")
	}

	klog.V(1).Infof("Tool call: argocd_get_application completed successfully in %v", duration)
	return NewTextResult(sb.String(), nil), nil
}

// argocdSyncApplication syncs an ArgoCD application
func (s *Server) argocdSyncApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_sync_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	revision := ctr.GetString("revision", "")
	prune := ctr.GetBool("prune", false)
	dryRun := ctr.GetBool("dry_run", false)
	namespace := ctr.GetString("namespace", "")

	// Extract resources array
	resources := ctr.GetStringSlice("resources", []string{})

	klog.V(1).Infof("Tool: argocd_sync_application - name: %s, namespace: %s, revision: %s, prune: %t, dry_run: %t, resources: %v - got called",
		name, namespace, revision, prune, dryRun, resources)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_sync_application failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Sync the application
	err = argoClient.SyncApplication(ctx, name, revision, prune, dryRun)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_sync_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to sync application '%s': %w", name, err)), nil
	}

	// Prepare the response message
	mode := "live"
	if dryRun {
		mode = "dry-run"
	}

	pruneMsg := ""
	if prune {
		pruneMsg = " with pruning enabled"
	}

	revisionMsg := ""
	if revision != "" {
		revisionMsg = fmt.Sprintf(" to revision '%s'", revision)
	}

	resourcesMsg := ""
	if len(resources) > 0 {
		resourcesMsg = fmt.Sprintf(" (specific resources: %s)", strings.Join(resources, ", "))
	}

	// Build a more detailed success message
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully initiated sync of application '%s'%s in %s mode%s%s\n\n",
		name, revisionMsg, mode, pruneMsg, resourcesMsg))

	// Get the latest application status to provide more context
	app, appErr := argoClient.GetApplication(ctx, name, "normal")
	if appErr == nil {
		sb.WriteString("Current application status:\n")
		sb.WriteString(fmt.Sprintf("- Sync Status: %s\n", app.Status.Sync.Status))
		sb.WriteString(fmt.Sprintf("- Health Status: %s\n", app.Status.Health.Status))

		if app.Status.OperationState != nil {
			sb.WriteString(fmt.Sprintf("- Operation Phase: %s\n", app.Status.OperationState.Phase))
			if app.Status.OperationState.Message != "" {
				sb.WriteString(fmt.Sprintf("- Operation Message: %s\n", app.Status.OperationState.Message))
			}
		}

		// Add resource status information
		if len(app.Status.Resources) > 0 {
			sb.WriteString("\nResources:\n")
			for _, res := range app.Status.Resources {
				sb.WriteString(fmt.Sprintf("- %s/%s: %s\n",
					res.Kind, res.Name, res.Status))
			}
		}
	}

	result := sb.String()
	klog.V(1).Infof("Tool call: argocd_sync_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdCreateApplication creates a new ArgoCD application
func (s *Server) argocdCreateApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
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
	namespace := ctr.GetString("namespace", "")
	validate := ctr.GetBool("validate", true)
	upsert := ctr.GetBool("upsert", false)

	klog.V(1).Infof("Tool: argocd_create_application - name: %s, project: %s, repo_url: %s, path: %s, dest_server: %s, dest_namespace: %s, revision: %s, automated_sync: %t, prune: %t, self_heal: %t, namespace: %s, validate: %t, upsert: %t - got called",
		name, project, repoURL, path, destServer, destNamespace, revision, automatedSync, prune, selfHeal, namespace, validate, upsert)

	// Parse boolean flags with defaults

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Create the application
	createdApp, err := argoClient.CreateApplication(ctx, name, project, repoURL, path, destServer, destNamespace,
		revision, automatedSync, prune, selfHeal, namespace, validate, upsert)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to create application '%s': %w", name, err)), nil
	}

	// Format the response with a clear success message
	successMsg := fmt.Sprintf("Successfully created application '%s' in project '%s'.\n\nApplication details:\n", name, project)

	// Format the application details as JSON
	appDetails, err := formatJSON(createdApp)
	if err != nil {
		klog.Errorf("Tool call: argocd_create_application failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	// Combine success message with application details
	result := successMsg + appDetails

	klog.V(1).Infof("Tool call: argocd_create_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdUpdateApplication updates an existing ArgoCD application
func (s *Server) argocdUpdateApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
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

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, "")
	if err != nil {
		klog.Errorf("Tool call: argocd_update_application failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Update the application
	updatedApp, err := argoClient.UpdateApplication(ctx, name, project, repoURL, path, destServer, destNamespace,
		revision, automatedSync, prune, selfHeal, validate)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_update_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to update application '%s': %w", name, err)), nil
	}

	// Create a success message listing what was updated
	var updatedFields []string
	if project != "" {
		updatedFields = append(updatedFields, "project")
	}
	if repoURL != "" {
		updatedFields = append(updatedFields, "repository URL")
	}
	if path != "" {
		updatedFields = append(updatedFields, "path")
	}
	if destServer != "" {
		updatedFields = append(updatedFields, "destination server")
	}
	if destNamespace != "" {
		updatedFields = append(updatedFields, "destination namespace")
	}
	if revision != "" {
		updatedFields = append(updatedFields, "revision")
	}
	if automatedSync != nil || prune != nil || selfHeal != nil {
		updatedFields = append(updatedFields, "sync policy")
	}

	var updateDesc string
	if len(updatedFields) > 0 {
		updateDesc = fmt.Sprintf(" with updated %s", strings.Join(updatedFields, ", "))
	}

	successMsg := fmt.Sprintf("Successfully updated application '%s'%s.\n\nApplication details:\n", name, updateDesc)

	// Format the updated application as JSON
	appDetails, err := formatJSON(updatedApp)
	if err != nil {
		klog.Errorf("Tool call: argocd_update_application failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	// Combine success message with application details
	result := successMsg + appDetails

	klog.V(1).Infof("Tool call: argocd_update_application completed successfully in %v", duration)
	return NewTextResult(result, nil), nil
}

// argocdDeleteApplication deletes an ArgoCD application
func (s *Server) argocdDeleteApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_delete_application failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	cascade := ctr.GetBool("cascade", true)
	propagationPolicy := ctr.GetString("propagation_policy", "")
	namespace := ctr.GetString("namespace", "")

	// Default cascade to true if not specified

	// Validate propagation policy
	if propagationPolicy != "" &&
		propagationPolicy != "foreground" &&
		propagationPolicy != "background" &&
		propagationPolicy != "orphan" {
		klog.Errorf("Tool call: argocd_delete_application failed after %v: invalid propagation policy: %s", time.Since(start), propagationPolicy)
		return NewTextResult("", fmt.Errorf("invalid propagation policy: must be 'foreground', 'background', or 'orphan'")), nil
	}

	klog.V(1).Infof("Tool: argocd_delete_application - name: %s, cascade: %t, propagation_policy: %s, namespace: %s - got called",
		name, cascade, propagationPolicy, namespace)

	// Log the operation for debugging
	klog.Infof("Deleting ArgoCD application '%s' (cascade=%t, propagationPolicy=%s, namespace=%s)",
		name, cascade, propagationPolicy, namespace)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_delete_application failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// First check if the application exists
	app, err := argoClient.GetApplication(ctx, name, "")
	if err != nil {
		duration := time.Since(start)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			klog.Errorf("Tool call: argocd_delete_application failed after %v: application not found", duration)
			return NewTextResult("", fmt.Errorf("application '%s' not found", name)), nil
		}
		klog.Errorf("Tool call: argocd_delete_application failed after %v: failed to verify application existence: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to verify application existence: %w", err)), nil
	}

	// Capture some details about the application before deleting it
	appProject := app.Spec.Project
	appDestNamespace := app.Spec.Destination.Namespace
	appRepoURL := app.Spec.Source.RepoURL
	appPath := app.Spec.Source.Path

	// Delete the application
	err = argoClient.DeleteApplication(ctx, name, cascade, propagationPolicy, namespace)
	duration := time.Since(start)

	if err != nil {
		if strings.Contains(err.Error(), "415") {
			// Specific handling for content type errors
			klog.Errorf("Tool call: argocd_delete_application failed after %v: content type error (415)", duration)
			return NewTextResult("", fmt.Errorf("content type error (415) when deleting application. This may indicate an API compatibility issue with ArgoCD server")), nil
		}
		klog.Errorf("Tool call: argocd_delete_application failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to delete application '%s': %w", name, err)), nil
	}

	// Build detailed success message
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ Successfully deleted ArgoCD application '%s'\n\n", name))
	sb.WriteString("Deleted application details:\n")
	sb.WriteString(fmt.Sprintf("- Project: %s\n", appProject))
	sb.WriteString(fmt.Sprintf("- Namespace: %s\n", appDestNamespace))
	sb.WriteString(fmt.Sprintf("- Repository: %s\n", appRepoURL))
	sb.WriteString(fmt.Sprintf("- Path: %s\n", appPath))

	sb.WriteString("\nDeletion options:\n")
	if cascade {
		sb.WriteString("- Cascade: true (application resources were also deleted)\n")
	} else {
		sb.WriteString("- Cascade: false (application resources were preserved)\n")
	}

	if propagationPolicy != "" {
		sb.WriteString(fmt.Sprintf("- Propagation Policy: %s\n", propagationPolicy))
	}

	klog.Info(fmt.Sprintf("Successfully deleted application '%s'", name))
	klog.V(1).Infof("Tool call: argocd_delete_application completed successfully in %v", duration)
	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationResourceTree returns the resource tree for an application
func (s *Server) argocdGetApplicationResourceTree(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_resource_tree failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: argocd_get_application_resource_tree - name: %s, namespace: %s - got called", name, namespace)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_resource_tree failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application resource tree
	resourceTree, err := argoClient.GetApplicationResourceTree(ctx, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_resource_tree failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get resource tree for application '%s': %w", name, err)), nil
	}

	// Format and return the resource tree as JSON
	jsonResult, err := formatJSON(resourceTree)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_resource_tree failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Resource tree for application '%s':\n\n", name))

	sb.WriteString("Full resource tree (JSON):\n")
	sb.WriteString(jsonResult)

	// Add summary of resource counts by kind
	if len(resourceTree.Nodes) > 0 {
		kindCounts := make(map[string]int)
		for _, node := range resourceTree.Nodes {
			kindCounts[node.Kind]++
		}

		sb.WriteString("Resource summary by kind:\n")
		for kind, count := range kindCounts {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", kind, count))
		}
		sb.WriteString("\n")
	}

	klog.V(1).Infof("Tool call: argocd_get_application_resource_tree completed successfully in %v - found %d nodes", duration, len(resourceTree.Nodes))
	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationManagedResources returns the managed resources for an application
func (s *Server) argocdGetApplicationManagedResources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_managed_resources failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: argocd_get_application_managed_resources - name: %s, namespace: %s - got called", name, namespace)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_managed_resources failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application managed resources
	managedResources, err := argoClient.GetApplicationManagedResources(ctx, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_managed_resources failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get managed resources for application '%s': %w", name, err)), nil
	}

	// Format and return the managed resources as JSON
	jsonResult, err := formatJSON(managedResources)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_managed_resources failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString("Full managed resources (JSON):\n")
	sb.WriteString(jsonResult)

	sb.WriteString(fmt.Sprintf("Managed resources for application '%s':\n\n", name))

	// Add summary of resource counts
	if len(managedResources.Items) > 0 {
		sb.WriteString(fmt.Sprintf("Total managed resources: %d\n\n", len(managedResources.Items)))

		// Group resources by kind and count them
		kindCounts := make(map[string]int)
		for _, item := range managedResources.Items {
			key := fmt.Sprintf("%s/%s", item.Group, item.Kind)
			if item.Group == "" {
				key = item.Kind
			}
			kindCounts[key]++
		}

		sb.WriteString("Resources by kind:\n")
		for kind, count := range kindCounts {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", kind, count))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("No managed resources found.\n\n")
	}

	klog.V(1).Infof("Tool call: argocd_get_application_managed_resources completed successfully in %v - found %d resources", duration, len(managedResources.Items))
	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationWorkloadLogs returns logs for application workload
func (s *Server) argocdGetApplicationWorkloadLogs(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("application_name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing application_name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, err := ctr.RequireString("application_namespace")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: missing application_namespace parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
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
			if err := json.Unmarshal([]byte(ref), &resourceRef); err != nil {
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

	klog.V(1).Infof("Tool: argocd_get_application_workload_logs - application: %s, namespace: %s, resource: %s/%s/%s, tail: %s, follow: %t - got called",
		name, appNamespace, resourceRef.Kind, resourceRef.Namespace, resourceRef.Name, tailStr, follow)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get workload logs
	logs, err := argoClient.GetWorkloadLogs(ctx, name, appNamespace, resourceRef, follow, tailStr)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_workload_logs failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get logs for workload '%s' in application '%s': %w",
			resourceRef.Name, name, err)), nil
	}

	// Format the logs
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Logs for %s '%s' in namespace '%s' from application '%s':\n\n",
		resourceRef.Kind, resourceRef.Name, resourceRef.Namespace, name))

	if resourceRef.Container != "" {
		sb.WriteString(fmt.Sprintf("Container: %s\n\n", resourceRef.Container))
	}

	if len(logs) == 0 {
		sb.WriteString("No logs available.\n")
	} else {
		for _, logEntry := range logs {
			sb.WriteString(logEntry.Content)
			sb.WriteString("\n")
		}
	}

	klog.V(1).Infof("Tool call: argocd_get_application_workload_logs completed successfully in %v - retrieved %d log entries", duration, len(logs))
	return NewTextResult(sb.String(), nil), nil
}

// argocdGetResourceEvents returns events for a resource managed by an application
func (s *Server) argocdGetResourceEvents(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("application_name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: missing application_name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, err := ctr.RequireString("application_namespace")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: missing application_namespace parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
	}

	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: failed to get arguments", time.Since(start))
		return NewTextResult("", fmt.Errorf("failed to get arguments")), nil
	}
	resourceRef := argsMap["resource_ref"].(map[string]interface{})

	resourceName, ok := resourceRef["name"].(string)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: missing resource_ref.name", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
	}

	resourceNamespace, ok := resourceRef["namespace"].(string)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: missing resource_ref.namespace", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
	}

	klog.V(1).Infof("Tool: argocd_get_resource_events - application: %s, app_namespace: %s, resource: %s/%s - got called",
		name, appNamespace, resourceNamespace, resourceName)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get resource events
	events, err := argoClient.GetResourceEvents(ctx, name, appNamespace, resourceNamespace, resourceName)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get events for resource '%s' in application '%s': %w",
			resourceName, name, err)), nil
	}

	// Format and return the events
	jsonResult, err := formatJSON(events)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_events failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Events for resource '%s' in namespace '%s' from application '%s':\n\n",
		resourceName, resourceNamespace, name))

	if len(events.Items) == 0 {
		sb.WriteString("No events found.\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Found %d events:\n\n", len(events.Items)))

		// Create a summary table with the most relevant event information
		sb.WriteString("TYPE      REASON               AGE     MESSAGE\n")
		sb.WriteString("----      ------               ---     -------\n")

		for _, event := range events.Items {
			// Format age as a duration from LastTimestamp to now
			age := "unknown"
			if event.LastTimestamp != "" {
				if lastTime, err := time.Parse(time.RFC3339, event.LastTimestamp); err == nil {
					age = formatTimeAgo(time.Since(lastTime))
				}
			}

			// Truncate message if it's too long
			message := event.Message
			if len(message) > 50 {
				message = message[:47] + "..."
			}

			sb.WriteString(fmt.Sprintf("%-9s %-20s %-8s %s\n",
				event.Type,
				event.Reason,
				age,
				message))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("Full events (JSON):\n")
	sb.WriteString(jsonResult)

	klog.V(1).Infof("Tool call: argocd_get_resource_events completed successfully in %v - found %d events", duration, len(events.Items))
	return NewTextResult(sb.String(), nil), nil
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
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, err := ctr.RequireString("app_namespace")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing app_namespace parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
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
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.name", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
	}

	resourceNamespace, ok := resourceRef["namespace"].(string)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.namespace", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
	}

	resourceKind, ok := resourceRef["kind"].(string)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.kind", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.kind is required")), nil
	}

	resourceGroup, ok := resourceRef["group"].(string)
	if !ok {
		resourceGroup = "" // group can be empty for core resources
	}

	resourceVersion, ok := resourceRef["version"].(string)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: missing resource_ref.version", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.version is required")), nil
	}

	klog.V(1).Infof("Tool: argocd_get_resource_actions - application: %s, app_namespace: %s, resource: %s/%s/%s/%s/%s - got called",
		name, appNamespace, resourceGroup, resourceVersion, resourceKind, resourceNamespace, resourceName)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get resource actions
	actions, err := argoClient.GetResourceActions(ctx, name, appNamespace, resourceNamespace, resourceName,
		resourceKind, resourceGroup, resourceVersion)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get actions for resource '%s' in application '%s': %w",
			resourceName, name, err)), nil
	}

	// Format and return the actions
	jsonResult, err := formatJSON(actions)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_resource_actions failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available actions for %s '%s' in namespace '%s' from application '%s':\n\n",
		resourceKind, resourceName, resourceNamespace, name))

	if len(actions.Actions) == 0 {
		sb.WriteString("No actions available for this resource.\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Found %d available actions:\n\n", len(actions.Actions)))

		for _, action := range actions.Actions {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", action.Name, action.DisplayName))
			if action.Disabled {
				sb.WriteString("  (Currently disabled)\n")
			}
			if action.Background {
				sb.WriteString("  (Runs in background)\n")
			}
		}

		sb.WriteString("\n")
	}

	sb.WriteString("Full actions (JSON):\n")
	sb.WriteString(jsonResult)

	klog.V(1).Infof("Tool call: argocd_get_resource_actions completed successfully in %v - found %d actions", duration, len(actions.Actions))
	return NewTextResult(sb.String(), nil), nil
}

// argocdRunResourceAction runs an action on a resource managed by an application
func (s *Server) argocdRunResourceAction(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, err := ctr.RequireString("app_namespace")
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing app_namespace parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
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
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing resource_ref.name", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.name is required")), nil
	}

	resourceNamespace, ok := resourceRef["namespace"].(string)
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: missing resource_ref.namespace", time.Since(start))
		return NewTextResult("", fmt.Errorf("resource_ref.namespace is required")), nil
	}

	resourceKind, ok := resourceRef["kind"].(string)
	if err != nil {
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

	klog.V(1).Infof("Tool: argocd_run_resource_action - application: %s, app_namespace: %s, resource: %s/%s/%s/%s/%s, action: %s - got called",
		name, appNamespace, resourceGroup, resourceVersion, resourceKind, resourceNamespace, resourceName, action)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Run the action
	result, err := argoClient.RunResourceAction(ctx, name, appNamespace, resourceNamespace, resourceName,
		resourceKind, resourceGroup, resourceVersion, action)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to run action '%s' on resource '%s' in application '%s': %w",
			action, resourceName, name, err)), nil
	}

	// Format and return the result
	jsonResult, err := formatJSON(result)
	if err != nil {
		klog.Errorf("Tool call: argocd_run_resource_action failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully ran action '%s' on %s '%s' in namespace '%s' from application '%s'.\n\n",
		action, resourceKind, resourceName, resourceNamespace, name))

	sb.WriteString("Action result (JSON):\n")
	sb.WriteString(jsonResult)

	klog.V(1).Infof("Tool call: argocd_run_resource_action completed successfully in %v", duration)
	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationEvents returns events for an application
func (s *Server) argocdGetApplicationEvents(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	// Extract parameters
	name, err := ctr.RequireString("application_name")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed after %v: missing application_name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace, err := ctr.RequireString("application_namespace")
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed after %v: missing application_namespace parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
	}

	klog.V(1).Infof("Tool: argocd_get_application_events - application: %s, namespace: %s - got called", name, namespace)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed after %v: failed to initialize ArgoCD client: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			klog.Warningf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application events
	events, err := argoClient.GetApplicationEvents(ctx, name, namespace)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get events for application '%s': %w", name, err)), nil
	}

	// Format as JSON with indentation
	jsonResult, err := formatJSON(events)
	if err != nil {
		klog.Errorf("Tool call: argocd_get_application_events failed after %v: failed to format JSON: %v", duration, err)
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Events for application '%s':\n\n", name))

	if len(events.Items) == 0 {
		sb.WriteString("No events found.\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Total events: %d\n\n", len(events.Items)))

		// Show a table-like structure for events
		sb.WriteString("REASON                 TYPE       AGE        MESSAGE\n")
		sb.WriteString("------                 ----       ---        -------\n")

		for _, event := range events.Items {
			// Format age as a duration from LastTimestamp to now
			age := "unknown"
			if event.LastTimestamp != "" {
				if lastTime, err := time.Parse(time.RFC3339, event.LastTimestamp); err == nil {
					age = formatTimeAgo(time.Since(lastTime))
				}
			}

			// Truncate message if it's too long
			message := event.Message
			if len(message) > 60 {
				message = message[:57] + "..."
			}

			sb.WriteString(fmt.Sprintf("%-22s %-10s %-10s %s\n",
				event.Reason,
				event.Type,
				age,
				message))
		}

		sb.WriteString("\n")

		// Add more detailed information for each event
		sb.WriteString("Event Details:\n\n")
		for i, event := range events.Items {
			sb.WriteString(fmt.Sprintf("Event #%d:\n", i+1))
			sb.WriteString(fmt.Sprintf("  Type: %s\n", event.Type))
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", event.Reason))
			sb.WriteString(fmt.Sprintf("  Message: %s\n", event.Message))
			if event.Count > 1 {
				sb.WriteString(fmt.Sprintf("  Count: %d\n", event.Count))
			}
			if event.FirstTimestamp != "" {
				sb.WriteString(fmt.Sprintf("  First Seen: %s\n", event.FirstTimestamp))
			}
			if event.LastTimestamp != "" {
				sb.WriteString(fmt.Sprintf("  Last Seen: %s\n", event.LastTimestamp))
			}
			if event.InvolvedObject.Kind != "" {
				sb.WriteString(fmt.Sprintf("  Involved Object: %s/%s\n", event.InvolvedObject.Kind, event.InvolvedObject.Name))
			}
			if event.Source.Component != "" {
				sb.WriteString(fmt.Sprintf("  Source: %s\n", event.Source.Component))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("Full events (JSON):\n")
	sb.WriteString(jsonResult)

	klog.V(1).Infof("Tool call: argocd_get_application_events completed successfully in %v - found %d events", duration, len(events.Items))
	return NewTextResult(sb.String(), nil), nil
}
