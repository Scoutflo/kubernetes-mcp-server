package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"
)

// HelmClient provides an interface to the Helm SDK
type HelmClient struct {
	settings     *cli.EnvSettings
	actionConfig *action.Configuration
}

// NewHelmClient creates a new Helm client using the Kubernetes config
func (k *Kubernetes) NewHelmClient(namespace string) (*HelmClient, error) {
	settings := cli.New()

	// Use the namespace provided or the default one
	if namespace != "" {
		settings.SetNamespace(namespace)
	} else {
		settings.SetNamespace(namespaceOrDefault(""))
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		// Optional debug logging can be enabled here
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action configuration: %w", err)
	}

	return &HelmClient{
		settings:     settings,
		actionConfig: actionConfig,
	}, nil
}

// AddRepository adds a Helm chart repository
func (h *HelmClient) AddRepository(ctx context.Context, name, url string) error {
	// Initialize the repository manager
	repoFile := h.settings.RepositoryConfig

	// Ensure the repository directory exists
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Initialize the repo file if it doesn't already exist
	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read repository file: %w", err)
	}

	var f repo.File
	if err == nil {
		if err := yaml.Unmarshal(b, &f); err != nil {
			return fmt.Errorf("failed to unmarshal repository file: %w", err)
		}
	}

	// Check if repository already exists
	for _, r := range f.Repositories {
		if r.Name == name {
			return fmt.Errorf("repository %q already exists", name)
		}
	}

	// Add the new repository
	entry := &repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(entry, getter.All(h.settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	// Download the repository index
	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("failed to download repository index: %w", err)
	}

	// Add the repository to the file and save it
	f.Update(entry)
	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	return nil
}

// ListRepositories returns a list of configured Helm repositories
func (h *HelmClient) ListRepositories(ctx context.Context) ([]*repo.Entry, error) {
	// Initialize the repository manager
	repoFile := h.settings.RepositoryConfig

	// Check if the repository file exists
	_, err := os.Stat(repoFile)
	if os.IsNotExist(err) {
		// No repositories configured yet
		return []*repo.Entry{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to access repository file: %w", err)
	}

	// Read the repository file
	b, err := os.ReadFile(repoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository file: %w", err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repository file: %w", err)
	}

	return f.Repositories, nil
}

// UpdateRepositories updates all Helm repositories to get the latest charts
// If repoNames is provided, only updates the specified repositories
func (h *HelmClient) UpdateRepositories(ctx context.Context, repoNames ...string) (string, error) {
	// Initialize the repository manager
	repoFile := h.settings.RepositoryConfig

	// Check if the repository file exists
	_, err := os.Stat(repoFile)
	if os.IsNotExist(err) {
		return "No repositories to update", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to access repository file: %w", err)
	}

	// Read the repository file
	b, err := os.ReadFile(repoFile)
	if err != nil {
		return "", fmt.Errorf("failed to read repository file: %w", err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return "", fmt.Errorf("failed to unmarshal repository file: %w", err)
	}

	// Create a map for quick lookups if specific repos are requested
	repoMap := make(map[string]bool)
	for _, name := range repoNames {
		repoMap[name] = true
	}

	// Update each repository
	var updatedRepos []string
	var notFoundRepos []string

	// If specific repos were requested, check if they exist
	if len(repoNames) > 0 {
		for _, name := range repoNames {
			found := false
			for _, entry := range f.Repositories {
				if entry.Name == name {
					found = true
					break
				}
			}
			if !found {
				notFoundRepos = append(notFoundRepos, name)
			}
		}

		if len(notFoundRepos) > 0 {
			return "", fmt.Errorf("repositories not found: %s", strings.Join(notFoundRepos, ", "))
		}
	}

	for _, entry := range f.Repositories {
		// If specific repos were requested, only update those
		if len(repoNames) > 0 && !repoMap[entry.Name] {
			continue
		}

		r, err := repo.NewChartRepository(entry, getter.All(h.settings))
		if err != nil {
			return "", fmt.Errorf("failed to create chart repository for %s: %w", entry.Name, err)
		}

		// Download the repository index
		if _, err := r.DownloadIndexFile(); err != nil {
			return "", fmt.Errorf("failed to update repository %s: %w", entry.Name, err)
		}

		updatedRepos = append(updatedRepos, entry.Name)
	}

	if len(updatedRepos) == 0 {
		return "No repositories to update", nil
	}

	if len(repoNames) > 0 {
		return fmt.Sprintf("Successfully updated repositories: %s", strings.Join(updatedRepos, ", ")), nil
	}

	return fmt.Sprintf("Successfully updated %d repositories", len(updatedRepos)), nil
}

// GetRelease retrieves specific information about a Helm release
// resource can be one of: all, hooks, manifest, notes, values
func (h *HelmClient) GetRelease(ctx context.Context, name, resource string) (string, error) {
	// Validate resource type
	validResources := map[string]bool{
		"all":      true,
		"hooks":    true,
		"manifest": true,
		"notes":    true,
		"values":   true,
	}

	if resource == "" {
		resource = "all" // Default to "all" if not specified
	}

	if !validResources[resource] {
		return "", fmt.Errorf("invalid resource type: %s, must be one of: all, hooks, manifest, notes, values", resource)
	}

	// Initialize the getter action
	client := action.NewGet(h.actionConfig)

	// Get the release
	rel, err := client.Run(name)
	if err != nil {
		return "", fmt.Errorf("failed to get release %s: %w", name, err)
	}

	// Return the requested information based on the resource type
	switch resource {
	case "all":
		return formatReleaseAll(rel)
	case "hooks":
		return formatReleaseHooks(rel)
	case "manifest":
		return rel.Manifest, nil
	case "notes":
		if rel.Info != nil {
			return rel.Info.Notes, nil
		}
		return "", nil
	case "values":
		return formatReleaseValues(rel)
	default:
		return "", fmt.Errorf("unknown resource type: %s", resource)
	}
}

// Helper function to format all release information
func formatReleaseAll(rel *release.Release) (string, error) {
	var sb strings.Builder

	// Basic release information
	sb.WriteString(fmt.Sprintf("NAME: %s\n", rel.Name))
	sb.WriteString(fmt.Sprintf("NAMESPACE: %s\n", rel.Namespace))

	if rel.Info != nil {
		sb.WriteString(fmt.Sprintf("LAST DEPLOYED: %s\n", rel.Info.LastDeployed.Format("Mon Jan 2 15:04:05 2006")))
		sb.WriteString(fmt.Sprintf("STATUS: %s\n", rel.Info.Status))
	}

	if rel.Chart != nil && rel.Chart.Metadata != nil {
		sb.WriteString(fmt.Sprintf("CHART: %s-%s\n", rel.Chart.Metadata.Name, rel.Chart.Metadata.Version))
		if rel.Chart.Metadata.AppVersion != "" {
			sb.WriteString(fmt.Sprintf("APP VERSION: %s\n", rel.Chart.Metadata.AppVersion))
		}
	}

	// Values
	valuesOutput, err := formatReleaseValues(rel)
	if err != nil {
		return "", err
	}
	sb.WriteString("\nVALUES:\n")
	sb.WriteString(valuesOutput)

	// Manifest
	sb.WriteString("\nMANIFEST:\n")
	sb.WriteString(rel.Manifest)

	// Notes if available
	if rel.Info != nil && rel.Info.Notes != "" {
		sb.WriteString("\nNOTES:\n")
		sb.WriteString(rel.Info.Notes)
	}

	// Hooks if available
	if len(rel.Hooks) > 0 {
		hooksOutput, err := formatReleaseHooks(rel)
		if err != nil {
			return "", err
		}
		sb.WriteString("\nHOOKS:\n")
		sb.WriteString(hooksOutput)
	}

	return sb.String(), nil
}

// Helper function to format release hooks
func formatReleaseHooks(rel *release.Release) (string, error) {
	if len(rel.Hooks) == 0 {
		return "No hooks defined", nil
	}

	var sb strings.Builder
	for _, hook := range rel.Hooks {
		sb.WriteString(fmt.Sprintf("---\n# Source: %s\n%s\n", hook.Path, hook.Manifest))
	}

	return sb.String(), nil
}

// Helper function to format release values
func formatReleaseValues(rel *release.Release) (string, error) {
	// Get computed values (combination of defaults and user values)
	values, err := yaml.Marshal(rel.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal release values: %w", err)
	}

	return string(values), nil
}

// ListOptions contains options for listing releases
type ListOptions struct {
	AllNamespaces bool
	All           bool
	Uninstalled   bool
	Uninstalling  bool
	Failed        bool
	Deployed      bool
	Pending       bool
	Filter        string
	Output        string
}

// ListReleases lists Helm releases with various filtering options
func (h *HelmClient) ListReleases(ctx context.Context, opts ListOptions) (string, error) {
	// Initialize the list action
	client := action.NewList(h.actionConfig)

	// Apply filters
	// By default, shows only deployed and failed releases
	if opts.All {
		client.All = true
	}
	if opts.Uninstalled {
		client.Uninstalled = true
	}
	if opts.Uninstalling {
		client.Uninstalling = true
	}
	if opts.Failed {
		client.Failed = true
	}
	if opts.Deployed {
		client.Deployed = true
	}
	if opts.Pending {
		client.Pending = true
	}

	// If no specific state filter is provided, use the default (deployed and failed)
	if !opts.All && !opts.Uninstalled && !opts.Uninstalling && !opts.Failed && !opts.Deployed && !opts.Pending {
		client.Deployed = true
		client.Failed = true
	}

	// Apply namespace filter - special case for AllNamespaces
	if opts.AllNamespaces {
		client.AllNamespaces = true
	}

	// Apply regex filter
	if opts.Filter != "" {
		client.Filter = opts.Filter
	}

	// Limit to 256 results to prevent overwhelming output
	client.Limit = 256

	// Get the list of releases
	releases, err := client.Run()
	if err != nil {
		return "", fmt.Errorf("failed to list releases: %w", err)
	}

	if len(releases) == 0 {
		return "No releases found", nil
	}

	// Format the output
	output := opts.Output
	if output == "" {
		output = "table" // Default to table format
	}

	switch output {
	case "json":
		return formatReleasesJSON(releases)
	case "yaml":
		return formatReleasesYAML(releases)
	default:
		return formatReleasesTable(releases)
	}
}

// Helper function to format releases as a table
func formatReleasesTable(releases []*release.Release) (string, error) {
	if len(releases) == 0 {
		return "No releases found", nil
	}

	var sb strings.Builder

	// Table header
	sb.WriteString("NAME\tNAMESPACE\tREVISION\tUPDATED\tSTATUS\tCHART\tAPP VERSION\n")

	// Sort releases by name
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Name < releases[j].Name
	})

	// Table rows
	for _, r := range releases {
		updated := r.Info.LastDeployed.Format("2006-01-02 15:04:05")
		chartVersion := fmt.Sprintf("%s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version)
		appVersion := r.Chart.Metadata.AppVersion
		if appVersion == "" {
			appVersion = "n/a"
		}

		sb.WriteString(fmt.Sprintf("%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
			r.Name,
			r.Namespace,
			r.Version,
			updated,
			r.Info.Status,
			chartVersion,
			appVersion))
	}

	return sb.String(), nil
}

// Helper function to format releases as JSON
func formatReleasesJSON(releases []*release.Release) (string, error) {
	// Create a simplified structure for JSON output
	type releaseInfo struct {
		Name       string `json:"name"`
		Namespace  string `json:"namespace"`
		Revision   int    `json:"revision"`
		Updated    string `json:"updated"`
		Status     string `json:"status"`
		Chart      string `json:"chart"`
		AppVersion string `json:"app_version"`
	}

	var releaseList []releaseInfo
	for _, r := range releases {
		updated := r.Info.LastDeployed.Format(time.RFC3339)
		chartVersion := fmt.Sprintf("%s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version)
		appVersion := r.Chart.Metadata.AppVersion
		if appVersion == "" {
			appVersion = "n/a"
		}

		releaseList = append(releaseList, releaseInfo{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   r.Version,
			Updated:    updated,
			Status:     string(r.Info.Status),
			Chart:      chartVersion,
			AppVersion: appVersion,
		})
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(releaseList, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal releases to JSON: %w", err)
	}

	return string(jsonData), nil
}

// Helper function to format releases as YAML
func formatReleasesYAML(releases []*release.Release) (string, error) {
	// Create a simplified structure for YAML output (same as for JSON)
	type releaseInfo struct {
		Name       string `json:"name" yaml:"name"`
		Namespace  string `json:"namespace" yaml:"namespace"`
		Revision   int    `json:"revision" yaml:"revision"`
		Updated    string `json:"updated" yaml:"updated"`
		Status     string `json:"status" yaml:"status"`
		Chart      string `json:"chart" yaml:"chart"`
		AppVersion string `json:"app_version" yaml:"appVersion"`
	}

	var releaseList []releaseInfo
	for _, r := range releases {
		updated := r.Info.LastDeployed.Format(time.RFC3339)
		chartVersion := fmt.Sprintf("%s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version)
		appVersion := r.Chart.Metadata.AppVersion
		if appVersion == "" {
			appVersion = "n/a"
		}

		releaseList = append(releaseList, releaseInfo{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   r.Version,
			Updated:    updated,
			Status:     string(r.Info.Status),
			Chart:      chartVersion,
			AppVersion: appVersion,
		})
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(releaseList)
	if err != nil {
		return "", fmt.Errorf("failed to marshal releases to YAML: %w", err)
	}

	return string(yamlData), nil
}

// UninstallReleaseOptions contains options for uninstalling a release
type UninstallReleaseOptions struct {
	DryRun bool
	Wait   bool
}

// UninstallRelease uninstalls a Helm release
func (h *HelmClient) UninstallRelease(ctx context.Context, name string, opts UninstallReleaseOptions) (string, error) {
	// Initialize the uninstall action
	client := action.NewUninstall(h.actionConfig)

	// Apply options
	client.DryRun = opts.DryRun
	client.Wait = opts.Wait

	// Uninstall the release
	res, err := client.Run(name)
	if err != nil {
		return "", fmt.Errorf("failed to uninstall release %s: %w", name, err)
	}

	// If it was a dry run, indicate this in the output
	if opts.DryRun {
		return fmt.Sprintf("DRY RUN: Release %q would be uninstalled", name), nil
	}

	// Format the result message
	if res != nil && res.Info != "" {
		return fmt.Sprintf("Release %q has been uninstalled\n%s", name, res.Info), nil
	}

	return fmt.Sprintf("Release %q has been uninstalled", name), nil
}

// InstallOptions contains options for installing a release
type InstallOptions struct {
	Namespace string
	Set       []string
	Values    []string
	RepoURL   string
	Version   string
	// Wait      bool
}

// InstallRelease installs a Helm chart
func (h *HelmClient) InstallRelease(ctx context.Context, name, chart string, opts InstallOptions) (string, error) {
	// Initialize the install action
	client := action.NewInstall(h.actionConfig)

	// Set the release name and namespace
	client.ReleaseName = name
	if opts.Namespace != "" {
		client.Namespace = opts.Namespace
	}

	// Set chart version if specified
	if opts.Version != "" {
		client.Version = opts.Version
	}

	// Set repository URL if specified
	if opts.RepoURL != "" {
		client.RepoURL = opts.RepoURL
	}

	// Get chart
	chartPath, err := client.ChartPathOptions.LocateChart(chart, h.settings)
	if err != nil {
		return "", fmt.Errorf("failed to locate chart %s: %w", chart, err)
	}

	// Load the chart
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to load chart: %w", err)
	}

	// Merge values
	values := chartRequested.Values
	if len(opts.Values) > 0 {
		for _, valuesFile := range opts.Values {
			currentMap := map[string]interface{}{}
			bytes, err := os.ReadFile(valuesFile)
			if err != nil {
				return "", fmt.Errorf("failed to read values file %s: %w", valuesFile, err)
			}
			if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
				return "", fmt.Errorf("failed to parse values file %s: %w", valuesFile, err)
			}
			values = mergeMaps(values, currentMap)
		}
	}

	// Apply --set values
	if len(opts.Set) > 0 {
		for _, value := range opts.Set {
			if err := strvals.ParseInto(value, values); err != nil {
				return "", fmt.Errorf("failed to parse --set value %q: %w", value, err)
			}
		}
	}

	// Run the installation
	rel, err := client.Run(chartRequested, values)
	if err != nil {
		return "", fmt.Errorf("failed to install chart: %w", err)
	}

	return fmt.Sprintf("Release %q has been installed. Happy Helming!\n\n%s", name, rel.Info.Notes), nil
}

// Helper function to merge two maps
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// UpgradeOptions contains options for upgrading a release
type UpgradeOptions struct {
	Namespace string
	Set       []string
	Values    []string
	RepoURL   string
	Version   string
	// Wait      bool
}

// UpgradeRelease upgrades a Helm release to a new version of a chart
func (h *HelmClient) UpgradeRelease(ctx context.Context, name, chart string, opts UpgradeOptions) (string, error) {
	// Initialize the upgrade action
	client := action.NewUpgrade(h.actionConfig)

	// Set namespace if specified
	if opts.Namespace != "" {
		client.Namespace = opts.Namespace
	}

	// Set chart version if specified
	if opts.Version != "" {
		client.Version = opts.Version
	}

	// Set repository URL if specified
	if opts.RepoURL != "" {
		client.RepoURL = opts.RepoURL
	}

	// Get chart
	chartPath, err := client.ChartPathOptions.LocateChart(chart, h.settings)
	if err != nil {
		return "", fmt.Errorf("failed to locate chart %s: %w", chart, err)
	}

	// Load the chart
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to load chart: %w", err)
	}

	// Get the current release to merge values
	histClient := action.NewHistory(h.actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(name); err != nil {
		return "", fmt.Errorf("failed to get release history: %w", err)
	}

	// Merge values
	values := chartRequested.Values
	if len(opts.Values) > 0 {
		for _, valuesFile := range opts.Values {
			currentMap := map[string]interface{}{}
			bytes, err := os.ReadFile(valuesFile)
			if err != nil {
				return "", fmt.Errorf("failed to read values file %s: %w", valuesFile, err)
			}
			if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
				return "", fmt.Errorf("failed to parse values file %s: %w", valuesFile, err)
			}
			values = mergeMaps(values, currentMap)
		}
	}

	// Apply --set values
	if len(opts.Set) > 0 {
		for _, value := range opts.Set {
			if err := strvals.ParseInto(value, values); err != nil {
				return "", fmt.Errorf("failed to parse --set value %q: %w", value, err)
			}
		}
	}

	// Run the upgrade
	rel, err := client.Run(name, chartRequested, values)
	if err != nil {
		return "", fmt.Errorf("failed to upgrade release: %w", err)
	}

	return fmt.Sprintf("Release %q has been upgraded. Happy Helming!\n\n%s", name, rel.Info.Notes), nil
}
