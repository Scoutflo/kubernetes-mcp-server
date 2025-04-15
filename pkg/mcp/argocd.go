package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

	// Format as JSON with indentation
	result, err := formatJSON(app)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
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

	result := fmt.Sprintf("Successfully started sync of application '%s'%s in %s mode%s%s",
		name, revisionMsg, mode, pruneMsg, resourcesMsg)

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

	// Format the response
	result, err := formatJSON(createdApp)
	if err != nil {
		return NewTextResult("", err), nil
	}

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

	// Format the response
	result, err := formatJSON(updatedApp)
	if err != nil {
		return NewTextResult("", err), nil
	}

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
	_, err = argoClient.GetApplication(ctx, name, "")
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return NewTextResult("", fmt.Errorf("application '%s' not found", name)), nil
		}
		return NewTextResult("", fmt.Errorf("failed to verify application existence: %w", err)), nil
	}

	// Delete the application
	err = argoClient.DeleteApplication(ctx, name, cascade, propagationPolicy, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "415") {
			// Specific handling for content type errors
			return NewTextResult("", fmt.Errorf("content type error (415) when deleting application. This may indicate an API compatibility issue with ArgoCD server")), nil
		}
		return NewTextResult("", fmt.Errorf("failed to delete application '%s': %w", name, err)), nil
	}

	// Prepare success message
	cascadeMsg := ""
	if cascade {
		cascadeMsg = " and its resources"
	}

	propagationMsg := ""
	if propagationPolicy != "" {
		propagationMsg = fmt.Sprintf(" using propagation policy '%s'", propagationPolicy)
	}

	result := fmt.Sprintf("Successfully deleted application '%s'%s%s", name, cascadeMsg, propagationMsg)
	log.Infof(result)

	return NewTextResult(result, nil), nil
}
