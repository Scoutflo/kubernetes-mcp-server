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
	log "github.com/sirupsen/logrus"
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
					mcp.Description("Filter applications by namespace (optional)"),
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
				),
				mcp.WithString("refresh",
					mcp.Description("Forces application reconciliation if set to 'hard' or 'normal' (optional)"),
				),
			),
			Handler: s.argocdGetApplication,
		},
		{
			Tool: mcp.NewTool("argocd_get_application_event",
				mcp.WithDescription("Returns events for an application"),
				mcp.WithString("application_name",
					mcp.Description("The name of the application"),
					mcp.Required(),
				),
				mcp.WithString("application_namespace",
					mcp.Description("The application namespace (optional)"),
				),
			),
			Handler: s.argocdGetApplicationEvent,
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
					mcp.Description("The application namespace (optional)"),
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
					mcp.Description("The application namespace (optional)"),
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
					mcp.Description("The application namespace (optional)"),
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
				mcp.WithString("resource_name",
					mcp.Description("The name of the resource"),
					mcp.Required(),
				),
				mcp.WithString("resource_namespace",
					mcp.Description("The namespace of the resource"),
					mcp.Required(),
				),
				mcp.WithString("resource_kind",
					mcp.Description("The kind of the resource (e.g., Deployment, StatefulSet, Pod)"),
					mcp.Required(),
				),
				mcp.WithString("resource_group",
					mcp.Description("The API group of the resource"),
					mcp.Required(),
				),
				mcp.WithString("resource_version",
					mcp.Description("The API version of the resource"),
					mcp.Required(),
				),
				mcp.WithString("container",
					mcp.Description("The container name (optional, will use first container if not specified)"),
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
				mcp.WithString("resource_namespace",
					mcp.Description("The namespace of the resource"),
					mcp.Required(),
				),
				mcp.WithString("resource_name",
					mcp.Description("The name of the resource"),
					mcp.Required(),
				),
			),
			Handler: s.argocdGetResourceEvents,
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
				mcp.WithString("resource_name",
					mcp.Description("The name of the resource"),
				),
				mcp.WithString("resource_namespace",
					mcp.Description("The namespace of the resource"),
				),
				mcp.WithString("resource_kind",
					mcp.Description("The kind of the resource (e.g., Deployment, StatefulSet, Pod)"),
				),
				mcp.WithString("resource_group",
					mcp.Description("The API group of the resource"),
				),
				mcp.WithString("resource_version",
					mcp.Description("The API version of the resource"),
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
				mcp.WithString("resource_name",
					mcp.Description("The name of the resource"),
					mcp.Required(),
				),
				mcp.WithString("resource_namespace",
					mcp.Description("The namespace of the resource"),
					mcp.Required(),
				),
				mcp.WithString("resource_kind",
					mcp.Description("The kind of the resource (e.g., Deployment, StatefulSet, Pod)"),
					mcp.Required(),
				),
				mcp.WithString("resource_group",
					mcp.Description("The API group of the resource"),
					mcp.Required(),
				),
				mcp.WithString("resource_version",
					mcp.Description("The API version of the resource"),
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
	// Extract parameters from the tool request
	project, _ := ctr.Params.Arguments["project"].(string)
	name, _ := ctr.Params.Arguments["name"].(string)
	repo, _ := ctr.Params.Arguments["repo"].(string)
	namespace, _ := ctr.Params.Arguments["namespace"].(string)
	refresh, _ := ctr.Params.Arguments["refresh"].(string)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Call the ListApplications method with individual parameters
	appList, err := argoClient.ListApplications(ctx, project, name, repo, namespace, refresh)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list ArgoCD applications: %w", err)), nil
	}

	// Format the output to be more user-friendly
	if len(appList.Items) == 0 {
		return NewTextResult("No applications found matching the criteria", nil), nil
	}

	// Create a more readable summary format
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d applications:\n\n", len(appList.Items)))
	sb.WriteString("NAME                    PROJECT        SYNC STATUS   HEALTH STATUS\n")
	sb.WriteString("----                    -------        -----------   -------------\n")

	for _, app := range appList.Items {
		syncStatus := "Unknown"
		if app.Status.Sync.Status != "" {
			syncStatus = app.Status.Sync.Status
		}

		healthStatus := "Unknown"
		if app.Status.Health.Status != "" {
			healthStatus = app.Status.Health.Status
		}

		sb.WriteString(fmt.Sprintf("%-24s %-14s %-14s %s\n",
			app.Metadata.Name,
			app.Spec.Project,
			syncStatus,
			healthStatus))
	}

	// Return summary table view by default
	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplication gets detailed information about a specific application
func (s *Server) argocdGetApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace, _ := ctr.Params.Arguments["namespace"].(string)
	refresh, _ := ctr.Params.Arguments["refresh"].(string)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application details
	app, err := argoClient.GetApplication(ctx, name, refresh)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get application '%s': %w", name, err)), nil
	}

	// Build a user-friendly summary of the application
	var sb strings.Builder

	// Provide full JSON for detailed information
	sb.WriteString("\nFull Application Details (JSON):\n")

	// Format as JSON with indentation
	jsonResult, err := formatJSON(app)
	if err != nil {
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

	return NewTextResult(sb.String(), nil), nil
}

// argocdSyncApplication syncs an ArgoCD application
func (s *Server) argocdSyncApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	revision, _ := ctr.Params.Arguments["revision"].(string)
	pruneStr, _ := ctr.Params.Arguments["prune"].(string)
	dryRunStr, _ := ctr.Params.Arguments["dry_run"].(string)
	namespace, _ := ctr.Params.Arguments["namespace"].(string)

	// Extract resources array
	var resources []string
	if resourcesRaw, ok := ctr.Params.Arguments["resources"].([]interface{}); ok {
		for _, r := range resourcesRaw {
			if resource, ok := r.(string); ok {
				resources = append(resources, resource)
			}
		}
	}

	// Parse boolean flags
	prune := strings.ToLower(pruneStr) == "true"
	dryRun := strings.ToLower(dryRunStr) == "true"

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Sync the application
	err = argoClient.SyncApplication(ctx, name, revision, prune, dryRun)
	if err != nil {
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
	return NewTextResult(result, nil), nil
}

// argocdCreateApplication creates a new ArgoCD application
func (s *Server) argocdCreateApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	project, ok := ctr.Params.Arguments["project"].(string)
	if !ok || project == "" {
		return NewTextResult("", fmt.Errorf("project name is required")), nil
	}

	repoURL, ok := ctr.Params.Arguments["repo_url"].(string)
	if !ok || repoURL == "" {
		return NewTextResult("", fmt.Errorf("repository URL is required")), nil
	}

	path, ok := ctr.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return NewTextResult("", fmt.Errorf("repository path is required")), nil
	}

	destServer, ok := ctr.Params.Arguments["dest_server"].(string)
	if !ok || destServer == "" {
		return NewTextResult("", fmt.Errorf("destination server is required")), nil
	}

	destNamespace, ok := ctr.Params.Arguments["dest_namespace"].(string)
	if !ok || destNamespace == "" {
		return NewTextResult("", fmt.Errorf("destination namespace is required")), nil
	}

	// Extract optional parameters
	revision, _ := ctr.Params.Arguments["revision"].(string)
	if revision == "" {
		revision = "HEAD" // Default revision
	}

	automatedSyncStr, _ := ctr.Params.Arguments["automated_sync"].(string)
	pruneStr, _ := ctr.Params.Arguments["prune"].(string)
	selfHealStr, _ := ctr.Params.Arguments["self_heal"].(string)
	namespace, _ := ctr.Params.Arguments["namespace"].(string)
	validateStr, _ := ctr.Params.Arguments["validate"].(string)
	upsertStr, _ := ctr.Params.Arguments["upsert"].(string)

	// Parse boolean flags with defaults
	automatedSync := strings.ToLower(automatedSyncStr) == "true"
	prune := strings.ToLower(pruneStr) == "true"
	selfHeal := strings.ToLower(selfHealStr) == "true"
	validate := validateStr == "" || strings.ToLower(validateStr) == "true" // Default true
	upsert := strings.ToLower(upsertStr) == "true"

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Create the application
	createdApp, err := argoClient.CreateApplication(ctx, name, project, repoURL, path, destServer, destNamespace,
		revision, automatedSync, prune, selfHeal, namespace, validate, upsert)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to create application '%s': %w", name, err)), nil
	}

	// Format the response with a clear success message
	successMsg := fmt.Sprintf("Successfully created application '%s' in project '%s'.\n\nApplication details:\n", name, project)

	// Format the application details as JSON
	appDetails, err := formatJSON(createdApp)
	if err != nil {
		return NewTextResult("", err), nil
	}

	// Combine success message with application details
	result := successMsg + appDetails

	return NewTextResult(result, nil), nil
}

// argocdUpdateApplication updates an existing ArgoCD application
func (s *Server) argocdUpdateApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	// Extract optional parameters
	project, _ := ctr.Params.Arguments["project"].(string)
	repoURL, _ := ctr.Params.Arguments["repo_url"].(string)
	path, _ := ctr.Params.Arguments["path"].(string)
	destServer, _ := ctr.Params.Arguments["dest_server"].(string)
	destNamespace, _ := ctr.Params.Arguments["dest_namespace"].(string)
	revision, _ := ctr.Params.Arguments["revision"].(string)
	validateStr, _ := ctr.Params.Arguments["validate"].(string)

	// Parse and convert boolean parameters that might be optional
	var automatedSync, prune, selfHeal *bool

	if automatedSyncStr, ok := ctr.Params.Arguments["automated_sync"].(string); ok && automatedSyncStr != "" {
		autoSyncVal := strings.ToLower(automatedSyncStr) == "true"
		automatedSync = &autoSyncVal
	}

	if pruneStr, ok := ctr.Params.Arguments["prune"].(string); ok && pruneStr != "" {
		pruneVal := strings.ToLower(pruneStr) == "true"
		prune = &pruneVal
	}

	if selfHealStr, ok := ctr.Params.Arguments["self_heal"].(string); ok && selfHealStr != "" {
		selfHealVal := strings.ToLower(selfHealStr) == "true"
		selfHeal = &selfHealVal
	}

	// Default validate to true if not specified
	validate := validateStr == "" || strings.ToLower(validateStr) == "true"

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, "")
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Update the application
	updatedApp, err := argoClient.UpdateApplication(ctx, name, project, repoURL, path, destServer, destNamespace,
		revision, automatedSync, prune, selfHeal, validate)
	if err != nil {
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
		return NewTextResult("", err), nil
	}

	// Combine success message with application details
	result := successMsg + appDetails

	return NewTextResult(result, nil), nil
}

// argocdDeleteApplication deletes an ArgoCD application
func (s *Server) argocdDeleteApplication(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	cascadeStr, _ := ctr.Params.Arguments["cascade"].(string)
	propagationPolicy, _ := ctr.Params.Arguments["propagation_policy"].(string)
	namespace, _ := ctr.Params.Arguments["namespace"].(string)

	// Default cascade to true if not specified
	cascade := cascadeStr == "" || strings.ToLower(cascadeStr) == "true"

	// Validate propagation policy
	if propagationPolicy != "" &&
		propagationPolicy != "foreground" &&
		propagationPolicy != "background" &&
		propagationPolicy != "orphan" {
		return NewTextResult("", fmt.Errorf("invalid propagation policy: must be 'foreground', 'background', or 'orphan'")), nil
	}

	// Log the operation for debugging
	log.Infof("Deleting ArgoCD application '%s' (cascade=%t, propagationPolicy=%s, namespace=%s)",
		name, cascade, propagationPolicy, namespace)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// First check if the application exists
	app, err := argoClient.GetApplication(ctx, name, "")
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return NewTextResult("", fmt.Errorf("application '%s' not found", name)), nil
		}
		return NewTextResult("", fmt.Errorf("failed to verify application existence: %w", err)), nil
	}

	// Capture some details about the application before deleting it
	appProject := app.Spec.Project
	appDestNamespace := app.Spec.Destination.Namespace
	appRepoURL := app.Spec.Source.RepoURL
	appPath := app.Spec.Source.Path

	// Delete the application
	err = argoClient.DeleteApplication(ctx, name, cascade, propagationPolicy, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "415") {
			// Specific handling for content type errors
			return NewTextResult("", fmt.Errorf("content type error (415) when deleting application. This may indicate an API compatibility issue with ArgoCD server")), nil
		}
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

	log.Info(fmt.Sprintf("Successfully deleted application '%s'", name))
	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationResourceTree returns the resource tree for an application
func (s *Server) argocdGetApplicationResourceTree(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace, _ := ctr.Params.Arguments["namespace"].(string)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application resource tree
	resourceTree, err := argoClient.GetApplicationResourceTree(ctx, name)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get resource tree for application '%s': %w", name, err)), nil
	}

	// Format and return the resource tree as JSON
	jsonResult, err := formatJSON(resourceTree)
	if err != nil {
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

	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationManagedResources returns the managed resources for an application
func (s *Server) argocdGetApplicationManagedResources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace, _ := ctr.Params.Arguments["namespace"].(string)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application managed resources
	managedResources, err := argoClient.GetApplicationManagedResources(ctx, name)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get managed resources for application '%s': %w", name, err)), nil
	}

	// Format and return the managed resources as JSON
	jsonResult, err := formatJSON(managedResources)
	if err != nil {
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

	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationWorkloadLogs returns logs for an application workload
func (s *Server) argocdGetApplicationWorkloadLogs(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["application_name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, ok := ctr.Params.Arguments["application_namespace"].(string)
	if !ok || appNamespace == "" {
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
	}

	// Extract resource parameters
	resourceName, ok := ctr.Params.Arguments["resource_name"].(string)
	if !ok || resourceName == "" {
		return NewTextResult("", fmt.Errorf("resource name is required")), nil
	}

	resourceNamespace, ok := ctr.Params.Arguments["resource_namespace"].(string)
	if !ok || resourceNamespace == "" {
		return NewTextResult("", fmt.Errorf("resource namespace is required")), nil
	}

	resourceKind, ok := ctr.Params.Arguments["resource_kind"].(string)
	if !ok || resourceKind == "" {
		return NewTextResult("", fmt.Errorf("resource kind is required")), nil
	}

	resourceGroup, ok := ctr.Params.Arguments["resource_group"].(string)
	if !ok {
		return NewTextResult("", fmt.Errorf("resource group is required")), nil
	}

	resourceVersion, ok := ctr.Params.Arguments["resource_version"].(string)
	if !ok || resourceVersion == "" {
		return NewTextResult("", fmt.Errorf("resource version is required")), nil
	}

	// Optional parameters
	container, _ := ctr.Params.Arguments["container"].(string)

	// Create resource reference
	resourceRef := kubernetes.ResourceRef{
		Group:     resourceGroup,
		Version:   resourceVersion,
		Kind:      resourceKind,
		Namespace: resourceNamespace,
		Name:      resourceName,
		Container: container,
	}

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get workload logs
	logs, err := argoClient.GetWorkloadLogs(ctx, name, appNamespace, resourceRef)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get logs for workload '%s' in application '%s': %w",
			resourceName, name, err)), nil
	}

	// Format the logs
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Logs for %s '%s' in namespace '%s' from application '%s':\n\n",
		resourceKind, resourceName, resourceNamespace, name))

	if container != "" {
		sb.WriteString(fmt.Sprintf("Container: %s\n\n", container))
	}

	if len(logs) == 0 {
		sb.WriteString("No logs available.\n")
	} else {
		for _, logEntry := range logs {
			sb.WriteString(logEntry.Content)
			sb.WriteString("\n")
		}
	}

	return NewTextResult(sb.String(), nil), nil
}

// argocdGetResourceEvents returns events for a resource managed by an application
func (s *Server) argocdGetResourceEvents(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, ok := ctr.Params.Arguments["app_namespace"].(string)
	if !ok || appNamespace == "" {
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
	}

	resourceNamespace, ok := ctr.Params.Arguments["resource_namespace"].(string)

	resourceName, ok := ctr.Params.Arguments["resource_name"].(string)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get resource events
	events, err := argoClient.GetResourceEvents(ctx, name, appNamespace, resourceNamespace, resourceName)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get events for resource '%s' in application '%s': %w",
			resourceName, name, err)), nil
	}

	// Format and return the events
	jsonResult, err := formatJSON(events)
	if err != nil {
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
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, ok := ctr.Params.Arguments["app_namespace"].(string)
	if !ok || appNamespace == "" {
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
	}

	// Extract resource parameters
	resourceName, ok := ctr.Params.Arguments["resource_name"].(string)
	if !ok || resourceName == "" {
		return NewTextResult("", fmt.Errorf("resource name is required")), nil
	}

	resourceNamespace, ok := ctr.Params.Arguments["resource_namespace"].(string)
	if !ok || resourceNamespace == "" {
		return NewTextResult("", fmt.Errorf("resource namespace is required")), nil
	}

	resourceKind, ok := ctr.Params.Arguments["resource_kind"].(string)
	if !ok || resourceKind == "" {
		return NewTextResult("", fmt.Errorf("resource kind is required")), nil
	}

	resourceGroup, ok := ctr.Params.Arguments["resource_group"].(string)
	if !ok {
		return NewTextResult("", fmt.Errorf("resource group is required")), nil
	}

	resourceVersion, ok := ctr.Params.Arguments["resource_version"].(string)
	if !ok || resourceVersion == "" {
		return NewTextResult("", fmt.Errorf("resource version is required")), nil
	}

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get resource actions
	actions, err := argoClient.GetResourceActions(ctx, name, appNamespace, resourceNamespace, resourceName,
		resourceKind, resourceGroup, resourceVersion)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get actions for resource '%s' in application '%s': %w",
			resourceName, name, err)), nil
	}

	// Format and return the actions
	jsonResult, err := formatJSON(actions)
	if err != nil {
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available actions for %s '%s' in namespace '%s' from application '%s':\n\n",
		resourceKind, resourceName, resourceNamespace, name))

	sb.WriteString("Full actions (JSON):\n")
	sb.WriteString(jsonResult)

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

	return NewTextResult(sb.String(), nil), nil
}

// argocdRunResourceAction runs an action on a resource managed by an application
func (s *Server) argocdRunResourceAction(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	appNamespace, ok := ctr.Params.Arguments["app_namespace"].(string)
	if !ok || appNamespace == "" {
		return NewTextResult("", fmt.Errorf("application namespace is required")), nil
	}

	// Extract resource parameters
	resourceName, ok := ctr.Params.Arguments["resource_name"].(string)
	if !ok || resourceName == "" {
		return NewTextResult("", fmt.Errorf("resource name is required")), nil
	}

	resourceNamespace, ok := ctr.Params.Arguments["resource_namespace"].(string)
	if !ok || resourceNamespace == "" {
		return NewTextResult("", fmt.Errorf("resource namespace is required")), nil
	}

	resourceKind, ok := ctr.Params.Arguments["resource_kind"].(string)
	if !ok || resourceKind == "" {
		return NewTextResult("", fmt.Errorf("resource kind is required")), nil
	}

	resourceGroup, ok := ctr.Params.Arguments["resource_group"].(string)
	if !ok {
		return NewTextResult("", fmt.Errorf("resource group is required")), nil
	}

	resourceVersion, ok := ctr.Params.Arguments["resource_version"].(string)
	if !ok || resourceVersion == "" {
		return NewTextResult("", fmt.Errorf("resource version is required")), nil
	}

	action, ok := ctr.Params.Arguments["action"].(string)
	if !ok || action == "" {
		return NewTextResult("", fmt.Errorf("action name is required")), nil
	}

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, appNamespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Run the action
	result, err := argoClient.RunResourceAction(ctx, name, appNamespace, resourceNamespace, resourceName,
		resourceKind, resourceGroup, resourceVersion, action)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to run action '%s' on resource '%s' in application '%s': %w",
			action, resourceName, name, err)), nil
	}

	// Format and return the result
	jsonResult, err := formatJSON(result)
	if err != nil {
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully ran action '%s' on %s '%s' in namespace '%s' from application '%s'.\n\n",
		action, resourceKind, resourceName, resourceNamespace, name))

	sb.WriteString("Action result (JSON):\n")
	sb.WriteString(jsonResult)

	return NewTextResult(sb.String(), nil), nil
}

// argocdGetApplicationEvent returns events for an application
func (s *Server) argocdGetApplicationEvent(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["application_name"].(string)
	if !ok || name == "" {
		return NewTextResult("", fmt.Errorf("application name is required")), nil
	}

	namespace, _ := ctr.Params.Arguments["application_namespace"].(string)

	// Create ArgoCD client
	argoClient, closer, err := s.k.NewArgoClient(ctx, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize ArgoCD client: %w", err)), nil
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close ArgoCD client connection: %v", err)
		}
	}()

	// Get application events
	events, err := argoClient.GetApplicationEvents(ctx, name, namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get events for application '%s': %w", name, err)), nil
	}

	// Format as JSON with indentation
	jsonResult, err := formatJSON(events)
	if err != nil {
		return NewTextResult("", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Events for application '%s':\n\n", name))

	if len(events.Items) == 0 {
		sb.WriteString("No events found.\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Total events: %d\n\n", len(events.Items)))

		// Show a table-like structure for events
		for i, event := range events.Items {
			sb.WriteString(fmt.Sprintf("Event #%d:\n", i+1))
			sb.WriteString(fmt.Sprintf("  Type: %s\n", event.Type))
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", event.Reason))
			sb.WriteString(fmt.Sprintf("  Message: %s\n", event.Message))
			if event.LastTimestamp != "" {
				sb.WriteString(fmt.Sprintf("  Last Timestamp: %s\n", event.LastTimestamp))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("Full events (JSON):\n")
	sb.WriteString(jsonResult)

	return NewTextResult(sb.String(), nil), nil
}
