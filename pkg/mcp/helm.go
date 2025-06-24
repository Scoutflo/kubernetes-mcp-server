package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
	"k8s.io/klog/v2"
)

func (s *Server) initHelm() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("helm_add_repository",
			mcp.WithDescription("Add a Helm chart repository"),
			mcp.WithString("name",
				mcp.Description("Repository name"),
				mcp.Required(),
			),
			mcp.WithString("url",
				mcp.Description("Repository URL"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to use for Helm operations (optional)"),
			),
		), Handler: s.helmAddRepository},

		{Tool: mcp.NewTool("helm_list_repositories",
			mcp.WithDescription("List all configured Helm repositories"),
			mcp.WithString("random_string",
				mcp.Description("Dummy parameter for no-parameter tools"),
				mcp.Required(),
			),
		), Handler: s.helmListRepositories},

		{Tool: mcp.NewTool("helm_update_repositories",
			mcp.WithDescription("Update Helm repositories to get the latest charts"),
			mcp.WithString("repo_name",
				mcp.Description("Optional name of the repository to update. If not provided, all repositories will be updated"),
			),
		), Handler: s.helmUpdateRepositories},

		{Tool: mcp.NewTool("helm_get_release",
			mcp.WithDescription("Get detailed information about a Helm release, available resources are: "+
				"all (download all information for a named release), "+
				"hooks (download all hooks for a named release), "+
				"manifest (download the manifest for a named release. The manifest is a YAML-formatted file containing the complete state of the release.), "+
				"notes (download the notes for a named release. The notes are a text document that contains information about the release.), "+
				"values (download the values for a named release. The values are a YAML-formatted file containing the values for the release.)"),
			mcp.WithString("name",
				mcp.Description("The name of the release"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("The namespace to get the release from (optional)"),
			),
			mcp.WithString("resource",
				mcp.Description("The resource to get information about. If not provided, all resources will be returned, can be one of: all, hooks, manifest, notes, values"),
			),
		), Handler: s.helmGetRelease},

		{Tool: mcp.NewTool("helm_list_releases",
			mcp.WithDescription("List all of the Helm releases for a specific namespace "+
				"If the --filter flag is provided, it will be treated as a filter. Filters are "+
				"regular expressions (Perl compatible) that are applied to the list of releases. "+
				"Only items that match the filter will be returned. "+
				"Usage: helm list --filter 'ara[a-z]+' "+
				"NAME                UPDATED                                  CHART "+
				"maudlin-arachnid    2020-06-18 14:17:46.125134977 +0000 UTC  alpine-0.1.0"),
			mcp.WithString("namespace",
				mcp.Description("The namespace to list the helm charts from (optional)"),
			),
			mcp.WithString("all_namespaces",
				mcp.Description("If 'true', list releases from all namespaces (accepted values: 'true', 'false')"),
			),
			mcp.WithString("all",
				mcp.Description("If 'true', show all releases without any filter applied (accepted values: 'true', 'false')"),
			),
			mcp.WithString("uninstalled",
				mcp.Description("If 'true', list uninstalled releases (accepted values: 'true', 'false')"),
			),
			mcp.WithString("uninstalling",
				mcp.Description("If 'true', list uninstalling releases (accepted values: 'true', 'false')"),
			),
			mcp.WithString("failed",
				mcp.Description("If 'true', list failed releases (accepted values: 'true', 'false')"),
			),
			mcp.WithString("deployed",
				mcp.Description("If 'true', list deployed releases (accepted values: 'true', 'false')"),
			),
			mcp.WithString("pending",
				mcp.Description("If 'true', list pending releases (accepted values: 'true', 'false')"),
			),
			mcp.WithString("filter",
				mcp.Description("A regular expression (Perl compatible). Any releases that match the expression will be included in the results"),
			),
			mcp.WithString("output",
				mcp.Description("The output format of the helm list command, one of: table, json, yaml. Prefer table for human readability"),
			),
		), Handler: s.helmListReleases},

		{Tool: mcp.NewTool("helm_install_release",
			mcp.WithDescription("Install a Helm chart. The chart argument can be either: a chart reference('example/mariadb'), "+
				"a path to a chart directory, a packaged chart, or a fully qualified URL. "+
				"For chart references, the latest version will be specified unless the '--version' flag is set."),
			mcp.WithString("name",
				mcp.Description("The name of the release"),
				mcp.Required(),
			),
			mcp.WithString("chart",
				mcp.Description("The chart to install (chart reference, a path to packaged chart, a path to an unpacked chart directory or URL)"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("The namespace to install the release in"),
			),
			mcp.WithArray("set",
				mcp.Description("A list of key-value pairs to set on the release (e.g., [\"key1=val1\", \"key2=val2\"])"),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
			mcp.WithArray("values",
				mcp.Description("A list of files to use as the value source (e.g., [\"myvalues.yaml\", \"override.yaml\"])"),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
			mcp.WithString("repo_url",
				mcp.Description("Chart repository url where to locate the requested chart"),
			),
			mcp.WithString("version",
				mcp.Description("Specify a version constraint for the chart version to use"),
			),
			// mcp.WithString("wait",
			// 	mcp.Description("If 'true', wait for the release to be installed (accepted values: 'true', 'false')"),
			// ),
		), Handler: s.helmInstallRelease},

		{Tool: mcp.NewTool("helm_uninstall_release",
			mcp.WithDescription("Uninstall a Helm release takes a release name and namespace as arguments "+
				"It removes all of the resources associated with the last release of the chart "+
				"as well as the release history, freeing it up for future use. "+
				"Use the '--dry-run' flag to see which releases will be uninstalled without actually "+
				"uninstalling them. "+
				"Usage: helm uninstall RELEASE_NAME [...] [flags]"),
			mcp.WithString("name",
				mcp.Description("The name of the release"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("The namespace to uninstall the release from"),
				mcp.Required(),
			),
			mcp.WithString("dry_run",
				mcp.Description("If 'true', show which releases will be uninstalled without actually uninstalling them (accepted values: 'true', 'false')"),
			),
			mcp.WithString("wait",
				mcp.Description("If 'true', wait for the release to be uninstalled (accepted values: 'true', 'false')"),
			),
		), Handler: s.helmUninstallRelease},

		{Tool: mcp.NewTool("helm_upgrade_release",
			mcp.WithDescription("Upgrade a release to a new version of a chart. The upgrade arguments must be a release and chart. The chart "+
				"argument can be either: a chart reference('example/mariadb'), a path to a chart directory, "+
				"a packaged chart, or a fully qualified URL. For chart references, the latest "+
				"version will be specified unless the '--version' flag is set."),
			mcp.WithString("name",
				mcp.Description("The name of the release"),
				mcp.Required(),
			),
			mcp.WithString("chart",
				mcp.Description("The chart to upgrade (chart reference, a path to packaged chart, a path to an unpacked chart directory or URL)"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("The namespace to upgrade the release in"),
			),
			mcp.WithArray("set",
				mcp.Description("A list of key-value pairs to set on the release (e.g., [\"key1=val1\", \"key2=val2\"])"),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
			mcp.WithArray("values",
				mcp.Description("A list of files to use as the value source (e.g., [\"myvalues.yaml\", \"override.yaml\"])"),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
			mcp.WithString("repo_url",
				mcp.Description("Chart repository url where to locate the requested chart"),
			),
			mcp.WithString("version",
				mcp.Description("Specify a version constraint for the chart version to use"),
			),
			// mcp.WithString("wait",
			// 	mcp.Description("If 'true', wait for the release to be upgraded (accepted values: 'true', 'false')"),
			// ),
		), Handler: s.helmUpgradeRelease},
	}
}

func (s *Server) helmAddRepository(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: helm_add_repository failed after %v: missing or invalid repository name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to add repository: missing or invalid repository name")), nil
	}

	url, err := ctr.RequireString("url")
	if err != nil {
		klog.Errorf("Tool call: helm_add_repository failed after %v: missing or invalid repository URL by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to add repository: missing or invalid repository URL")), nil
	}

	klog.V(1).Infof("Tool: helm_add_repository - name: %s, url: %s - got called by session id: %s", name, url, sessionID)

	// Add repository using kubernetes client
	result, err := s.k.AddRepository(ctx, name, url)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_add_repository failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to add repository: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: helm_add_repository completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(result, nil), nil
}

func (s *Server) helmListRepositories(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: helm_list_repositories - got called by session id: %s", sessionID)

	// List repositories using kubernetes client
	repos, err := s.k.ListRepositories(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_list_repositories failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list repositories: %v", err)), nil
	}

	if len(repos) == 0 {
		klog.V(1).Infof("Tool call: helm_list_repositories completed successfully in %v, repositories_count: 0 by session id: %s", duration, sessionID)
		return NewTextResult("No Helm repositories configured", nil), nil
	}

	// Format the repositories list
	var sb strings.Builder
	sb.WriteString("Configured Helm repositories:\n\n")
	sb.WriteString("NAME                    URL\n")
	sb.WriteString("----                    ---\n")

	for _, repo := range repos {
		// Format with padding similar to helm repo list output
		sb.WriteString(fmt.Sprintf("%-23s %s\n", repo.Name, repo.URL))
	}

	klog.V(1).Infof("Tool call: helm_list_repositories completed successfully in %v, repositories_count: %d by session id: %s", duration, len(repos), sessionID)
	return NewTextResult(sb.String(), nil), nil
}

func (s *Server) helmUpdateRepositories(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Check if specific repo was specified
	var repoNames []string
	repoName := ctr.GetString("repo_name", "")
	if repoName != "" {
		repoNames = []string{repoName}
	}

	klog.V(1).Infof("Tool: helm_update_repositories - repo_name: %s, repos_to_update: %d - got called by session id: %s", repoName, len(repoNames), sessionID)

	// Update repositories using kubernetes client
	result, err := s.k.UpdateRepositories(ctx, repoNames...)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_update_repositories failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to update repositories: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: helm_update_repositories completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(result, nil), nil
}

func (s *Server) helmGetRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Extract parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: helm_get_release failed after %v: missing or invalid release name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to get release: missing or invalid release name")), nil
	}

	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: helm_get_release - name: %s, namespace: %s - got called by session id: %s", name, namespace, sessionID)

	// Get release information using kubernetes client
	result, err := s.k.GetRelease(ctx, name, namespace)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_get_release failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get release: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: helm_get_release completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

func (s *Server) helmListReleases(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Build the list options
	opts := kubernetes.ListOptions{}

	// Boolean options (as strings)
	opts.AllNamespaces = strings.ToLower(ctr.GetString("all_namespaces", "false")) == "true"

	opts.All = strings.ToLower(ctr.GetString("all", "false")) == "true"

	opts.Uninstalled = strings.ToLower(ctr.GetString("uninstalled", "false")) == "true"

	opts.Uninstalling = strings.ToLower(ctr.GetString("uninstalling", "false")) == "true"

	opts.Failed = strings.ToLower(ctr.GetString("failed", "false")) == "true"

	opts.Deployed = strings.ToLower(ctr.GetString("deployed", "false")) == "true"

	opts.Pending = strings.ToLower(ctr.GetString("pending", "false")) == "true"

	// String options
	opts.Filter = ctr.GetString("filter", "")

	opts.Output = ctr.GetString("output", "")

	klog.V(1).Infof("Tool: helm_list_releases - all_namespaces: %t, all: %t, filter: %s, output: %s - got called by session id: %s",
		opts.AllNamespaces, opts.All, opts.Filter, opts.Output, sessionID)

	// List releases using kubernetes client
	result, err := s.k.ListReleases(ctx, opts)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_list_releases failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list releases: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: helm_list_releases completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

func (s *Server) helmUninstallRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: helm_uninstall_release failed after %v: missing or invalid release name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to uninstall release: missing or invalid release name")), nil
	}

	namespace := ctr.GetString("namespace", "")
	if namespace == "" {
		klog.Errorf("Tool call: helm_uninstall_release failed after %v: missing or invalid namespace by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to uninstall release: missing or invalid namespace")), nil
	}

	// Build the uninstall options
	opts := kubernetes.UninstallReleaseOptions{}

	// Boolean options (as strings)
	opts.DryRun = strings.ToLower(ctr.GetString("dry_run", "false")) == "true"

	opts.Wait = strings.ToLower(ctr.GetString("wait", "false")) == "true"

	klog.V(1).Infof("Tool: helm_uninstall_release - name: %s, namespace: %s, dry_run: %t, wait: %t - got called by session id: %s",
		name, namespace, opts.DryRun, opts.Wait, sessionID)

	// Uninstall the release using kubernetes client
	result, err := s.k.UninstallRelease(ctx, name, namespace, opts)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_uninstall_release failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to uninstall release: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: helm_uninstall_release completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(result, nil), nil
}

func (s *Server) helmInstallRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: helm_install_release failed after %v: missing or invalid release name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to install release: missing or invalid release name")), nil
	}

	chart, err := ctr.RequireString("chart")
	if err != nil {
		klog.Errorf("Tool call: helm_install_release failed after %v: missing or invalid chart by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to install release: missing or invalid chart")), nil
	}

	// Extract optional parameters
	namespace := ctr.GetString("namespace", "")

	// Build the install options
	opts := kubernetes.InstallOptions{
		Namespace: namespace,
	}

	// Set repository URL if provided
	opts.RepoURL = ctr.GetString("repo_url", "")

	// Set version if provided
	opts.Version = ctr.GetString("version", "")

	// Array options
	args := ctr.GetRawArguments()
	if argsMap, ok := args.(map[string]interface{}); ok {
		if setValues, ok := argsMap["set"].([]interface{}); ok {
			for _, sv := range setValues {
				if strVal, ok := sv.(string); ok {
					opts.Set = append(opts.Set, strVal)
				}
			}
		}

		if valueFiles, ok := argsMap["values"].([]interface{}); ok {
			for _, vf := range valueFiles {
				if strVal, ok := vf.(string); ok {
					opts.Values = append(opts.Values, strVal)
				}
			}
		}
	}

	klog.V(1).Infof("Tool: helm_install_release - name: %s, chart: %s, namespace: %s, repo_url: %s, version: %s, set_count: %d, values_count: %d - got called by session id: %s",
		name, chart, namespace, opts.RepoURL, opts.Version, len(opts.Set), len(opts.Values), sessionID)

	// Install the release using kubernetes client
	result, err := s.k.InstallRelease(ctx, name, chart, opts)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_install_release failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to install release: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: helm_install_release completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

func (s *Server) helmUpgradeRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	// Extract required parameters
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: helm_upgrade_release failed after %v: missing or invalid release name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to upgrade release: missing or invalid release name")), nil
	}

	chart, err := ctr.RequireString("chart")
	if err != nil {
		klog.Errorf("Tool call: helm_upgrade_release failed after %v: missing or invalid chart by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to upgrade release: missing or invalid chart")), nil
	}

	// Extract optional parameters
	namespace := ctr.GetString("namespace", "")

	// Build the upgrade options
	opts := kubernetes.UpgradeOptions{
		Namespace: namespace,
	}

	// Set version if provided
	opts.Version = ctr.GetString("version", "")

	// Array options
	args := ctr.GetRawArguments()
	if argsMap, ok := args.(map[string]interface{}); ok {
		if setValues, ok := argsMap["set"].([]interface{}); ok {
			for _, sv := range setValues {
				if strVal, ok := sv.(string); ok {
					opts.Set = append(opts.Set, strVal)
				}
			}
		}

		if valueFiles, ok := argsMap["values"].([]interface{}); ok {
			for _, vf := range valueFiles {
				if strVal, ok := vf.(string); ok {
					opts.Values = append(opts.Values, strVal)
				}
			}
		}
	}

	klog.V(1).Infof("Tool: helm_upgrade_release - name: %s, chart: %s, namespace: %s, version: %s, set_count: %d, values_count: %d - got called by session id: %s",
		name, chart, namespace, opts.Version, len(opts.Set), len(opts.Values), sessionID)

	// Upgrade the release using kubernetes client
	result, err := s.k.UpgradeRelease(ctx, name, chart, opts)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: helm_upgrade_release failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to upgrade release: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: helm_upgrade_release completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}
