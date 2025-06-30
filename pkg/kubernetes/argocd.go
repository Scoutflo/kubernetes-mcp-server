package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ArgoCD API paths
const (
	ArgoCDSessionPath         = "/api/v1/session"
	ArgoCDApplicationsPath    = "/api/v1/applications"
	ArgoCDProjectsPath        = "/api/v1/projects"
	ArgoCDRepositoriesPath    = "/api/v1/repositories"
	ArgoCDApplicationPath     = "/api/v1/applications/%s"      // Requires app name
	ArgoCDApplicationSyncPath = "/api/v1/applications/%s/sync" // Requires app name
)

// ArgoCD connection config - this should be moved to a proper config manager
// TODO: Move these to a secure configuration store

// ArgoCD session request/response structures
type SessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionResponse struct {
	Token string `json:"token"`
}

// ArgoClient represents a client for ArgoCD REST API
type ArgoClient struct {
	serverURL  string
	authToken  string
	httpClient *http.Client
	namespace  string
}

// ApplicationList represents a list of applications from ArgoCD
type ApplicationList struct {
	Items []Application `json:"items"`
}

// Application represents an ArgoCD application
type Application struct {
	Kind       string            `json:"kind"`
	APIVersion string            `json:"apiVersion"`
	Metadata   Metadata          `json:"metadata"`
	Spec       ApplicationSpec   `json:"spec"`
	Status     ApplicationStatus `json:"status,omitempty"`
}

// Metadata contains application metadata
type Metadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty"`
	UID               string            `json:"uid,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

// ApplicationSpec contains the application specification
type ApplicationSpec struct {
	Source      ApplicationSource      `json:"source"`
	Destination ApplicationDestination `json:"destination"`
	Project     string                 `json:"project"`
	SyncPolicy  *SyncPolicy            `json:"syncPolicy,omitempty"`
}

// ApplicationSource contains information about application source
type ApplicationSource struct {
	RepoURL        string    `json:"repoURL"`
	Path           string    `json:"path,omitempty"`
	TargetRevision string    `json:"targetRevision,omitempty"`
	Chart          string    `json:"chart,omitempty"`
	Helm           *HelmSpec `json:"helm,omitempty"`
}

// HelmSpec contains Helm-specific options
type HelmSpec struct {
	Parameters []HelmParameter `json:"parameters,omitempty"`
	Values     string          `json:"values,omitempty"`
	ValueFiles []string        `json:"valueFiles,omitempty"`
}

// HelmParameter is a Helm parameter
type HelmParameter struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	ForceString bool   `json:"forceString,omitempty"`
}

// ApplicationDestination contains information about application destination
type ApplicationDestination struct {
	Server    string `json:"server,omitempty"`
	Namespace string `json:"namespace"`
	Name      string `json:"name,omitempty"`
}

// SyncPolicy contains sync policy configuration
type SyncPolicy struct {
	Automated                *Automated                `json:"automated,omitempty"`
	SyncOptions              []string                  `json:"syncOptions,omitempty"`
	Retry                    *RetryPolicy              `json:"retry,omitempty"`
	ManagedNamespaceMetadata *ManagedNamespaceMetadata `json:"managedNamespaceMetadata,omitempty"`
}

// Automated defines automated sync policy
type Automated struct {
	Prune      bool `json:"prune,omitempty"`
	SelfHeal   bool `json:"selfHeal,omitempty"`
	AllowEmpty bool `json:"allowEmpty,omitempty"`
}

// RetryPolicy defines retry policy
type RetryPolicy struct {
	Limit   int64    `json:"limit,omitempty"`
	Backoff *Backoff `json:"backoff,omitempty"`
}

// Backoff defines backoff strategy
type Backoff struct {
	Duration    string `json:"duration,omitempty"`
	Factor      int64  `json:"factor,omitempty"`
	MaxDuration string `json:"maxDuration,omitempty"`
}

// ManagedNamespaceMetadata defines metadata to be applied to managed namespace
type ManagedNamespaceMetadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ApplicationStatus contains application status information
type ApplicationStatus struct {
	Sync           SyncStatus         `json:"sync,omitempty"`
	Health         HealthStatus       `json:"health,omitempty"`
	Resources      []ResourceStatus   `json:"resources,omitempty"`
	OperationState *OperationState    `json:"operationState,omitempty"`
	Summary        ApplicationSummary `json:"summary,omitempty"`
	ReconciledAt   string             `json:"reconciledAt,omitempty"`
	SourceType     string             `json:"sourceType,omitempty"`
}

// SyncStatus contains sync status information
type SyncStatus struct {
	Status   string `json:"status"`
	Revision string `json:"revision,omitempty"`
}

// HealthStatus contains application health status
type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ResourceStatus contains resource status information
type ResourceStatus struct {
	Group     string `json:"group,omitempty"`
	Version   string `json:"version,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Status    string `json:"status,omitempty"`
}

// OperationState contains information about operation state
type OperationState struct {
	Operation  Operation `json:"operation,omitempty"`
	Phase      string    `json:"phase,omitempty"`
	Message    string    `json:"message,omitempty"`
	StartedAt  string    `json:"startedAt,omitempty"`
	FinishedAt string    `json:"finishedAt,omitempty"`
}

// Operation defines a sync operation
type Operation struct {
	Sync *SyncOperation `json:"sync,omitempty"`
}

// SyncOperation defines sync operation details
type SyncOperation struct {
	Revision    string   `json:"revision,omitempty"`
	Prune       bool     `json:"prune,omitempty"`
	DryRun      bool     `json:"dryRun,omitempty"`
	SyncOptions []string `json:"syncOptions,omitempty"`
}

// ApplicationSummary provides a summary of application resources
type ApplicationSummary struct {
	ExternalURLs []string `json:"externalURLs,omitempty"`
	Images       []string `json:"images,omitempty"`
}

// SyncRequest represents a request to sync an application
type SyncRequest struct {
	Name      string   `json:"name,omitempty"`
	Revision  string   `json:"revision,omitempty"`
	Prune     bool     `json:"prune,omitempty"`
	DryRun    bool     `json:"dryRun,omitempty"`
	Strategy  string   `json:"strategy,omitempty"`
	Resources []string `json:"resources,omitempty"`
}

// ProjectList represents a list of ArgoCD projects
type ProjectList struct {
	Items []Project `json:"items"`
}

// Project represents an ArgoCD project
type Project struct {
	Kind       string        `json:"kind"`
	APIVersion string        `json:"apiVersion"`
	Metadata   Metadata      `json:"metadata"`
	Spec       ProjectSpec   `json:"spec"`
	Status     ProjectStatus `json:"status,omitempty"`
}

// ProjectSpec contains the project specification
type ProjectSpec struct {
	Description                string           `json:"description,omitempty"`
	SourceRepos                []string         `json:"sourceRepos,omitempty"`
	Destinations               []AppDestination `json:"destinations,omitempty"`
	ClusterResourceWhitelist   []ResourceRef    `json:"clusterResourceWhitelist,omitempty"`
	NamespaceResourceBlacklist []ResourceRef    `json:"namespaceResourceBlacklist,omitempty"`
	NamespaceResourceWhitelist []ResourceRef    `json:"namespaceResourceWhitelist,omitempty"`
	ClusterResourceBlacklist   []ResourceRef    `json:"clusterResourceBlacklist,omitempty"`
	Roles                      []ProjectRole    `json:"roles,omitempty"`
	SyncWindows                []SyncWindow     `json:"syncWindows,omitempty"`
}

// AppDestination contains information about project destination
type AppDestination struct {
	Server    string `json:"server,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

// ResourceRef reference to a Kubernetes resource
type ResourceRef struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Container string `json:"container,omitempty"`
}

// ProjectRole represents a user's role within a project
type ProjectRole struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Policies    []string `json:"policies,omitempty"`
	Groups      []string `json:"groups,omitempty"`
}

// SyncWindow represents a time window for syncing
type SyncWindow struct {
	Kind         string   `json:"kind"`
	Schedule     string   `json:"schedule"`
	Duration     string   `json:"duration"`
	Applications []string `json:"applications,omitempty"`
	Namespaces   []string `json:"namespaces,omitempty"`
	Clusters     []string `json:"clusters,omitempty"`
	ManualSync   bool     `json:"manualSync,omitempty"`
}

// ProjectStatus contains project status information
type ProjectStatus struct {
	JWTTokensByRole map[string]JWTToken `json:"jwtTokensByRole,omitempty"`
}

// JWTToken contains information about a JWT token
type JWTToken struct {
	IssuedAt int64 `json:"iat"`
}

// ListApplications lists ArgoCD applications with filtering options
func (k *Kubernetes) ListApplications(ctx context.Context, project, name, repo, refresh string) (string, error) {
	endpoint := "/apis/v1/argocd/applications"

	params := url.Values{}
	if project != "" {
		params.Add("project", project)
	}
	if name != "" {
		params.Add("name", name)
	}
	if repo != "" {
		params.Add("repo", repo)
	}
	if refresh != "" {
		params.Add("refresh", refresh)
	}

	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list ArgoCD applications: %w", err)
	}
	return string(response), nil
}

// GetApplication gets detailed information about a specific ArgoCD application
func (k *Kubernetes) GetApplication(ctx context.Context, name, refresh string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("application name is required")
	}

	endpoint := "/apis/v1/argocd/application"

	params := url.Values{}
	params.Add("name", name)
	if refresh != "" {
		params.Add("refresh", refresh)
	}

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get ArgoCD application %s: %w", name, err)
	}
	return string(response), nil
}

// GetApplicationEvents returns events for an ArgoCD application
func (k *Kubernetes) GetApplicationEvents(ctx context.Context, appName string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("application name is required")
	}

	endpoint := "/apis/v1/argocd/application-events"

	params := url.Values{}
	params.Add("name", appName)

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get events for ArgoCD application %s: %w", appName, err)
	}
	return string(response), nil
}

// SyncApplication syncs an ArgoCD application
func (k *Kubernetes) SyncApplication(ctx context.Context, name, revision string, prune, dryRun bool) (string, error) {
	if name == "" {
		return "", fmt.Errorf("application name is required")
	}

	endpoint := "/apis/v1/argocd/sync-application"

	params := url.Values{}
	params.Add("name", name)
	if revision != "" {
		params.Add("revision", revision)
	}
	if prune {
		params.Add("prune", "true")
	}
	if dryRun {
		params.Add("dry_run", "true")
	}

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to sync ArgoCD application %s: %w", name, err)
	}
	return string(response), nil
}

// CreateApplicationRequest represents the request body for creating an application
type CreateApplicationRequest struct {
	Name          string `json:"name"`
	Project       string `json:"project"`
	RepoURL       string `json:"repo_url"`
	Path          string `json:"path"`
	DestServer    string `json:"dest_server"`
	DestNamespace string `json:"dest_namespace"`
	Revision      string `json:"revision,omitempty"`
	AutomatedSync string `json:"automated_sync,omitempty"`
	Prune         string `json:"prune,omitempty"`
	SelfHeal      string `json:"self_heal,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	Validate      string `json:"validate,omitempty"`
	Upsert        string `json:"upsert,omitempty"`
}

// CreateApplication creates a new ArgoCD application
func (k *Kubernetes) CreateApplication(ctx context.Context, name, project, repoURL, path, destServer, destNamespace, revision, automatedSync, prune, selfHeal, validate, upsert string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("application name is required")
	}
	if project == "" {
		return "", fmt.Errorf("project name is required")
	}

	endpoint := "/apis/v1/argocd/create-application"

	requestBody := CreateApplicationRequest{
		Name:          name,
		Project:       project,
		RepoURL:       repoURL,
		Path:          path,
		DestServer:    destServer,
		DestNamespace: destNamespace,
		Revision:      revision,
		AutomatedSync: automatedSync,
		Prune:         prune,
		SelfHeal:      selfHeal,
		Validate:      validate,
		Upsert:        upsert,
	}

	response, err := k.MakeAPIRequest("POST", endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create ArgoCD application %s: %w", name, err)
	}
	return string(response), nil
}

// UpdateApplicationRequest represents the request body for updating an application
type UpdateApplicationRequest struct {
	Name          string `json:"name"`
	Project       string `json:"project,omitempty"`
	RepoURL       string `json:"repo_url,omitempty"`
	Path          string `json:"path,omitempty"`
	DestServer    string `json:"dest_server,omitempty"`
	DestNamespace string `json:"dest_namespace,omitempty"`
	Revision      string `json:"revision,omitempty"`
	AutomatedSync string `json:"automated_sync,omitempty"`
	Prune         string `json:"prune,omitempty"`
	SelfHeal      string `json:"self_heal,omitempty"`
	Validate      string `json:"validate,omitempty"`
}

// UpdateApplication updates an existing ArgoCD application
func (k *Kubernetes) UpdateApplication(ctx context.Context, name, project, repoURL, path, destServer, destNamespace, revision, automatedSync, prune, selfHeal, validate string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("application name is required")
	}

	endpoint := "/apis/v1/argocd/update-application"

	requestBody := UpdateApplicationRequest{
		Name:          name,
		Project:       project,
		RepoURL:       repoURL,
		Path:          path,
		DestServer:    destServer,
		DestNamespace: destNamespace,
		Revision:      revision,
		AutomatedSync: automatedSync,
		Prune:         prune,
		SelfHeal:      selfHeal,
		Validate:      validate,
	}

	response, err := k.MakeAPIRequest("POST", endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to update ArgoCD application %s: %w", name, err)
	}
	return string(response), nil
}

// DeleteApplicationRequest represents the request body for deleting an application
type DeleteApplicationRequest struct {
	Name              string `json:"name"`
	Cascade           string `json:"cascade,omitempty"`
	PropagationPolicy string `json:"propagation_policy,omitempty"`
}

// DeleteApplication deletes an ArgoCD application
func (k *Kubernetes) DeleteApplication(ctx context.Context, name, cascade, propagationPolicy string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("application name is required")
	}

	endpoint := "/apis/v1/argocd/delete-application"

	requestBody := DeleteApplicationRequest{
		Name:              name,
		Cascade:           cascade,
		PropagationPolicy: propagationPolicy,
	}

	response, err := k.MakeAPIRequest("POST", endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to delete ArgoCD application %s: %w", name, err)
	}
	return string(response), nil
}

// GetApplicationResourceTree gets the resource tree for an ArgoCD application
func (k *Kubernetes) GetApplicationResourceTree(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("application name is required")
	}

	endpoint := "/apis/v1/argocd/application-resource-tree"

	params := url.Values{}
	params.Add("name", name)

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get resource tree for ArgoCD application %s: %w", name, err)
	}
	return string(response), nil
}

// GetApplicationManagedResources gets the managed resources for an ArgoCD application
func (k *Kubernetes) GetApplicationManagedResources(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("application name is required")
	}

	endpoint := "/apis/v1/argocd/application-managed-resources"

	params := url.Values{}
	params.Add("name", name)

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get managed resources for ArgoCD application %s: %w", name, err)
	}
	return string(response), nil
}

// GetApplicationWorkloadLogs gets logs for application workload
func (k *Kubernetes) GetApplicationWorkloadLogs(ctx context.Context, appName, resourceRef, tail, follow string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("application name is required")
	}
	if resourceRef == "" {
		return "", fmt.Errorf("resource reference is required")
	}

	endpoint := "/apis/v1/argocd/application-workload-logs"

	params := url.Values{}
	params.Add("application_name", appName)
	params.Add("resource_ref", resourceRef)
	if tail != "" {
		params.Add("tail", tail)
	}
	if follow != "" {
		params.Add("follow", follow)
	}

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get workload logs for ArgoCD application %s: %w", appName, err)
	}
	return string(response), nil
}

// GetApplicationResourceEvents gets events for a resource managed by an application
func (k *Kubernetes) GetApplicationResourceEvents(ctx context.Context, appName, resourceRef string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("application name is required")
	}
	if resourceRef == "" {
		return "", fmt.Errorf("resource reference is required")
	}

	endpoint := "/apis/v1/argocd/application-resource-events"

	params := url.Values{}
	params.Add("application_name", appName)
	params.Add("resource_ref", resourceRef)

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get resource events for ArgoCD application %s: %w", appName, err)
	}
	return string(response), nil
}

// GetResourceActions gets available actions for a resource managed by an application
func (k *Kubernetes) GetResourceActions(ctx context.Context, appName, resourceRef string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("application name is required")
	}
	if resourceRef == "" {
		return "", fmt.Errorf("resource reference is required")
	}

	endpoint := "/apis/v1/argocd/resource-actions"

	params := url.Values{}
	params.Add("application_name", appName)
	params.Add("resource_ref", resourceRef)

	endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get resource actions for ArgoCD application %s: %w", appName, err)
	}
	return string(response), nil
}

// RunResourceActionRequest represents the request body for running a resource action
type RunResourceActionRequest struct {
	ApplicationName string `json:"application_name"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceRef     string `json:"resource_ref"`
	Action          string `json:"action"`
}

// RunResourceAction runs an action on a resource managed by an application
func (k *Kubernetes) RunResourceAction(ctx context.Context, appName, resourceRef, action string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("application name is required")
	}
	if resourceRef == "" {
		return "", fmt.Errorf("resource reference is required")
	}
	if action == "" {
		return "", fmt.Errorf("action is required")
	}

	endpoint := "/apis/v1/argocd/run-resource-action"

	requestBody := RunResourceActionRequest{
		ApplicationName: appName,
		ResourceRef:     resourceRef,
		Action:          action,
	}

	response, err := k.MakeAPIRequest("POST", endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to run resource action for ArgoCD application %s: %w", appName, err)
	}
	return string(response), nil
}
