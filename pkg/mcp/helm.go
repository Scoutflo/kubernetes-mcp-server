package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
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
			),
			mcp.WithArray("values",
				mcp.Description("A list of files to use as the value source (e.g., [\"myvalues.yaml\", \"override.yaml\"])"),
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
			),
			mcp.WithArray("values",
				mcp.Description("A list of files to use as the value source (e.g., [\"myvalues.yaml\", \"override.yaml\"])"),
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
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", errors.New("failed to add repository: missing or invalid repository name")), nil
	}

	url, ok := ctr.Params.Arguments["url"].(string)
	if !ok || url == "" {
		return NewTextResult("", errors.New("failed to add repository: missing or invalid repository URL")), nil
	}

	namespace := ""
	if ns, ok := ctr.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// Create Helm client
	helmClient, err := s.k.NewHelmClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// Add repository
	if err := helmClient.AddRepository(ctx, name, url); err != nil {
		return NewTextResult("", fmt.Errorf("failed to add repository: %v", err)), nil
	}

	result := fmt.Sprintf("Successfully added Helm repository '%s' with URL '%s'", name, url)
	return NewTextResult(result, nil), nil
}

func (s *Server) helmListRepositories(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create Helm client with default namespace
	helmClient, err := s.k.NewHelmClient("")
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// List repositories
	repos, err := helmClient.ListRepositories(ctx)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list repositories: %v", err)), nil
	}

	if len(repos) == 0 {
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

	return NewTextResult(sb.String(), nil), nil
}

func (s *Server) helmUpdateRepositories(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create Helm client with default namespace
	helmClient, err := s.k.NewHelmClient("")
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// Check if specific repo was specified
	var repoNames []string
	if repoName, ok := ctr.Params.Arguments["repo_name"].(string); ok && repoName != "" {
		repoNames = []string{repoName}
	}

	// Update repositories
	result, err := helmClient.UpdateRepositories(ctx, repoNames...)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to update repositories: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}

func (s *Server) helmGetRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", errors.New("failed to get release: missing or invalid release name")), nil
	}

	namespace := ""
	if ns, ok := ctr.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	resource := ""
	if res, ok := ctr.Params.Arguments["resource"].(string); ok {
		resource = res
	}

	// Create Helm client
	helmClient, err := s.k.NewHelmClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// Get release information
	result, err := helmClient.GetRelease(ctx, name, resource)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get release: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}

func (s *Server) helmListReleases(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	namespace := ""
	if ns, ok := ctr.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// Build the list options
	opts := kubernetes.ListOptions{}

	// Boolean options (as strings)
	if val, ok := ctr.Params.Arguments["all_namespaces"].(string); ok {
		opts.AllNamespaces = strings.ToLower(val) == "true"
	}

	if val, ok := ctr.Params.Arguments["all"].(string); ok {
		opts.All = strings.ToLower(val) == "true"
	}

	if val, ok := ctr.Params.Arguments["uninstalled"].(string); ok {
		opts.Uninstalled = strings.ToLower(val) == "true"
	}

	if val, ok := ctr.Params.Arguments["uninstalling"].(string); ok {
		opts.Uninstalling = strings.ToLower(val) == "true"
	}

	if val, ok := ctr.Params.Arguments["failed"].(string); ok {
		opts.Failed = strings.ToLower(val) == "true"
	}

	if val, ok := ctr.Params.Arguments["deployed"].(string); ok {
		opts.Deployed = strings.ToLower(val) == "true"
	}

	if val, ok := ctr.Params.Arguments["pending"].(string); ok {
		opts.Pending = strings.ToLower(val) == "true"
	}

	// String options
	if val, ok := ctr.Params.Arguments["filter"].(string); ok {
		opts.Filter = val
	}

	if val, ok := ctr.Params.Arguments["output"].(string); ok {
		opts.Output = val
	}

	// Create Helm client
	helmClient, err := s.k.NewHelmClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// List releases
	result, err := helmClient.ListReleases(ctx, opts)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list releases: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}

func (s *Server) helmUninstallRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", errors.New("failed to uninstall release: missing or invalid release name")), nil
	}

	namespace, ok := ctr.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		return NewTextResult("", errors.New("failed to uninstall release: missing or invalid namespace")), nil
	}

	// Build the uninstall options
	opts := kubernetes.UninstallReleaseOptions{}

	// Boolean options (as strings)
	if dryRun, ok := ctr.Params.Arguments["dry_run"].(string); ok {
		opts.DryRun = strings.ToLower(dryRun) == "true"
	}

	if wait, ok := ctr.Params.Arguments["wait"].(string); ok {
		opts.Wait = strings.ToLower(wait) == "true"
	}

	// Create Helm client
	helmClient, err := s.k.NewHelmClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// Uninstall the release
	result, err := helmClient.UninstallRelease(ctx, name, opts)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to uninstall release: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}

func (s *Server) helmInstallRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", errors.New("failed to install release: missing or invalid release name")), nil
	}

	chart, ok := ctr.Params.Arguments["chart"].(string)
	if !ok || chart == "" {
		return NewTextResult("", errors.New("failed to install release: missing or invalid chart")), nil
	}

	// Extract optional parameters
	namespace := ""
	if ns, ok := ctr.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// Build the install options
	opts := kubernetes.InstallOptions{
		Namespace: namespace,
	}

	// Set repository URL if provided
	if repoURL, ok := ctr.Params.Arguments["repo_url"].(string); ok {
		opts.RepoURL = repoURL
	}

	// Set version if provided
	if version, ok := ctr.Params.Arguments["version"].(string); ok {
		opts.Version = version
	}

	// Array options
	if setValues, ok := ctr.Params.Arguments["set"].([]interface{}); ok {
		for _, sv := range setValues {
			if strVal, ok := sv.(string); ok {
				opts.Set = append(opts.Set, strVal)
			}
		}
	}

	if valueFiles, ok := ctr.Params.Arguments["values"].([]interface{}); ok {
		for _, vf := range valueFiles {
			if strVal, ok := vf.(string); ok {
				opts.Values = append(opts.Values, strVal)
			}
		}
	}

	// Create Helm client
	helmClient, err := s.k.NewHelmClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// Install the release
	result, err := helmClient.InstallRelease(ctx, name, chart, opts)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to install release: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}

func (s *Server) helmUpgradeRelease(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	name, ok := ctr.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return NewTextResult("", errors.New("failed to upgrade release: missing or invalid release name")), nil
	}

	chart, ok := ctr.Params.Arguments["chart"].(string)
	if !ok || chart == "" {
		return NewTextResult("", errors.New("failed to upgrade release: missing or invalid chart")), nil
	}

	// Extract optional parameters
	namespace := ""
	if ns, ok := ctr.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// Build the upgrade options
	opts := kubernetes.UpgradeOptions{
		Namespace: namespace,
	}

	// Set version if provided
	if version, ok := ctr.Params.Arguments["version"].(string); ok {
		opts.Version = version
	}

	// Array options
	if setValues, ok := ctr.Params.Arguments["set"].([]interface{}); ok {
		for _, sv := range setValues {
			if strVal, ok := sv.(string); ok {
				opts.Set = append(opts.Set, strVal)
			}
		}
	}

	if valueFiles, ok := ctr.Params.Arguments["values"].([]interface{}); ok {
		for _, vf := range valueFiles {
			if strVal, ok := vf.(string); ok {
				opts.Values = append(opts.Values, strVal)
			}
		}
	}

	// Create Helm client
	helmClient, err := s.k.NewHelmClient(namespace)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to initialize Helm client: %v", err)), nil
	}

	// Upgrade the release
	result, err := helmClient.UpgradeRelease(ctx, name, chart, opts)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to upgrade release: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}
