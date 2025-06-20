package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// AddRepository adds a Helm chart repository via API call
func (k *Kubernetes) AddRepository(ctx context.Context, name, url string) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"name": name,
		"url":  url,
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "api/v1/helm/repositories", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to add repository: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	return string(response), nil
}

// ListRepositories returns a list of configured Helm repositories via API call
func (k *Kubernetes) ListRepositories(ctx context.Context) ([]*RepoEntry, error) {
	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", "api/v1/helm/repositories", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %v", err)
	}

	var result []*RepoEntry
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return result, nil
}

// RepoEntry represents a Helm repository entry
type RepoEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// UpdateRepositories updates all Helm repositories via API call
func (k *Kubernetes) UpdateRepositories(ctx context.Context, repoNames ...string) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"repoNames": repoNames,
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("PUT", "api/v1/helm/repositories/update", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to update repositories: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	return string(response), nil
}

// GetRelease retrieves specific information about a Helm release via API call
func (k *Kubernetes) GetRelease(ctx context.Context, name, namespace string) (string, error) {
	// Use default namespace if not provided
	if namespace == "" {
		namespace = "default"
	}
	endpoint := fmt.Sprintf("api/v1/helm/releases/%s/%s", namespace, name)

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get release: %v", err)
	}

	return string(response), nil
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

// ListReleases lists Helm releases via API call
func (k *Kubernetes) ListReleases(ctx context.Context, opts ListOptions) (string, error) {
	queryParams := url.Values{}

	// Add namespace if not all namespaces
	if !opts.AllNamespaces {
		namespace := "default"
		queryParams.Add("namespace", namespace)
	}

	if opts.AllNamespaces {
		queryParams.Add("all_namespaces", "true")
	}
	if opts.All {
		queryParams.Add("all", "true")
	}
	if opts.Uninstalled {
		queryParams.Add("uninstalled", "true")
	}
	if opts.Uninstalling {
		queryParams.Add("uninstalling", "true")
	}
	if opts.Failed {
		queryParams.Add("failed", "true")
	}
	if opts.Deployed {
		queryParams.Add("deployed", "true")
	}
	if opts.Pending {
		queryParams.Add("pending", "true")
	}
	if opts.Filter != "" {
		queryParams.Add("filter", opts.Filter)
	}
	if opts.Output != "" {
		queryParams.Add("output", opts.Output)
	}

	endpoint := "api/v1/helm/releases"
	if len(queryParams) > 0 {
		endpoint += "?" + queryParams.Encode()
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list releases: %v", err)
	}

	return string(response), nil
}

// UninstallReleaseOptions contains options for uninstalling a release
type UninstallReleaseOptions struct {
	DryRun bool
	Wait   bool
}

// UninstallRelease uninstalls a Helm release via API call
func (k *Kubernetes) UninstallRelease(ctx context.Context, name, namespace string, opts UninstallReleaseOptions) (string, error) {
	// Use default namespace if not provided
	if namespace == "" {
		namespace = "default"
	}
	queryParams := url.Values{}
	if opts.DryRun {
		queryParams.Add("dry_run", "true")
	}
	if opts.Wait {
		queryParams.Add("wait", "true")
	}

	endpoint := fmt.Sprintf("api/v1/helm/releases/%s/%s", namespace, name)
	if len(queryParams) > 0 {
		endpoint += "?" + queryParams.Encode()
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to uninstall release: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	// also want to return the result as a string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}

	return string(resultJSON), nil
}

// InstallOptions contains options for installing a release
type InstallOptions struct {
	Namespace string
	Set       []string
	Values    []string
	RepoURL   string
	Version   string
}

// InstallRelease installs a Helm chart via API call
func (k *Kubernetes) InstallRelease(ctx context.Context, name, chart string, opts InstallOptions) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"releaseName": name,
		"chartName":   chart,
		"namespace":   opts.Namespace,
	}

	if opts.RepoURL != "" {
		requestPayload["repoURL"] = opts.RepoURL
	}
	if opts.Version != "" {
		requestPayload["version"] = opts.Version
	}
	if len(opts.Set) > 0 {
		requestPayload["set"] = opts.Set
	}
	if len(opts.Values) > 0 {
		requestPayload["values"] = opts.Values
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "api/v1/helm/install", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to install release: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	// also want to return the result as a string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}

	return string(resultJSON), nil
}

// UpgradeOptions contains options for upgrading a release
type UpgradeOptions struct {
	Namespace string
	Set       []string
	Values    []string
	RepoURL   string
	Version   string
}

// UpgradeRelease upgrades a Helm release via API call
func (k *Kubernetes) UpgradeRelease(ctx context.Context, name, chart string, opts UpgradeOptions) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"releaseName": name,
		"chartName":   chart,
		"namespace":   opts.Namespace,
	}

	if opts.RepoURL != "" {
		requestPayload["repoURL"] = opts.RepoURL
	}
	if opts.Version != "" {
		requestPayload["version"] = opts.Version
	}
	if len(opts.Set) > 0 {
		requestPayload["set"] = opts.Set
	}
	if len(opts.Values) > 0 {
		requestPayload["values"] = opts.Values
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("PUT", "api/v1/helm/upgrade", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to upgrade release: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	// also want to return the result as a string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}

	return string(resultJSON), nil
}
