package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
var (
	argoServerURL = "https://argocd-sf-test-pp-sf9-i.scoutflo.agency/"
	argoApiToken  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJhZG1pbjphcGlLZXkiLCJuYmYiOjE3NDY2ODQ1NDIsImlhdCI6MTc0NjY4NDU0MiwianRpIjoiNDBiM2FhMDktMzVjZS00MTJjLWJlZDItNWYwMjVlYzkxODgxIn0.Rsi78M4kfyxtf2msQd0T-OroOtR_AuEfIAMB7xvrGNM"
)

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

// NewArgoClient creates a new ArgoCD HTTP client
func (k *Kubernetes) NewArgoClient(ctx context.Context, requestNamespace string) (*ArgoClient, io.Closer, error) {
	client := &ArgoClient{
		serverURL: argoServerURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		namespace: namespaceOrDefault(requestNamespace),
		authToken: argoApiToken, // Use the API token directly
	}

	// Create a closer function that does nothing since we don't have a persistent connection
	closer := io.NopCloser(strings.NewReader(""))

	return client, closer, nil
}

// doRequest is a helper function to make HTTP requests to ArgoCD API
func (c *ArgoClient) doRequest(ctx context.Context, method, path string, queryParams map[string]string, body interface{}) (*http.Response, error) {
	// Build full URL with query parameters
	apiURL := fmt.Sprintf("%s%s", strings.TrimSuffix(c.serverURL, "/"), path)

	if len(queryParams) > 0 {
		values := url.Values{}
		for k, v := range queryParams {
			values.Add(k, v)
		}
		apiURL = fmt.Sprintf("%s?%s", apiURL, values.Encode())
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add auth header if token exists
	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	}

	// Only set Content-Type for requests with a body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// For DELETE requests, explicitly add Accept header
	if method == http.MethodDelete {
		req.Header.Set("Accept", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	return resp, nil
}

// ListApplications lists ArgoCD applications with filtering
func (c *ArgoClient) ListApplications(ctx context.Context, project, name, repo, namespace, refresh string) (*ApplicationList, error) {
	// Build query parameters
	queryParams := make(map[string]string)
	if project != "" {
		queryParams["project"] = project
	}
	if name != "" {
		queryParams["name"] = name
	}
	if repo != "" {
		queryParams["repo"] = repo
	}
	if namespace != "" {
		queryParams["appNamespace"] = namespace
	}
	if refresh != "" {
		queryParams["refresh"] = refresh
	}

	resp, err := c.doRequest(ctx, http.MethodGet, ArgoCDApplicationsPath, queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list applications failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var appList ApplicationList
	err = json.NewDecoder(resp.Body).Decode(&appList)
	if err != nil {
		return nil, fmt.Errorf("error parsing applications list: %w", err)
	}

	return &appList, nil
}

// GetApplication gets details of a specific ArgoCD application
func (c *ArgoClient) GetApplication(ctx context.Context, appName string, refresh string) (*Application, error) {
	path := fmt.Sprintf(ArgoCDApplicationPath, appName)

	queryParams := make(map[string]string)
	if refresh != "" {
		queryParams["refresh"] = refresh
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get application failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var app Application
	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		return nil, fmt.Errorf("error parsing application: %w", err)
	}

	return &app, nil
}

// SyncApplication syncs an ArgoCD application
func (c *ArgoClient) SyncApplication(ctx context.Context, appName, revision string, prune, dryRun bool) error {
	path := fmt.Sprintf(ArgoCDApplicationSyncPath, appName)

	syncRequest := SyncRequest{
		Revision: revision,
		Prune:    prune,
		DryRun:   dryRun,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, path, nil, syncRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sync application failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListProjects lists ArgoCD projects
func (c *ArgoClient) ListProjects(ctx context.Context) (*ProjectList, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, ArgoCDProjectsPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list projects failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var projectList ProjectList
	err = json.NewDecoder(resp.Body).Decode(&projectList)
	if err != nil {
		return nil, fmt.Errorf("error parsing projects list: %w", err)
	}

	return &projectList, nil
}

// GetProject gets details of a specific ArgoCD project
func (c *ArgoClient) GetProject(ctx context.Context, projectName string) (*Project, error) {
	path := fmt.Sprintf("%s/%s", ArgoCDProjectsPath, projectName)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get project failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var project Project
	err = json.NewDecoder(resp.Body).Decode(&project)
	if err != nil {
		return nil, fmt.Errorf("error parsing project: %w", err)
	}

	return &project, nil
}

// CreateApplication creates a new ArgoCD application
func (c *ArgoClient) CreateApplication(ctx context.Context, name, project, repoURL, path, destServer, destNamespace, revision string,
	automatedSync, prune, selfHeal bool, namespace string, validate, upsert bool) (*Application, error) {

	// Build application object
	app := Application{
		Kind:       "Application",
		APIVersion: "argoproj.io/v1alpha1",
		Metadata: Metadata{
			Name: name,
		},
		Spec: ApplicationSpec{
			Project: project,
			Source: ApplicationSource{
				RepoURL:        repoURL,
				Path:           path,
				TargetRevision: revision,
			},
			Destination: ApplicationDestination{
				Server:    destServer,
				Namespace: destNamespace,
			},
		},
	}

	// Set namespace if provided
	if namespace != "" {
		app.Metadata.Namespace = namespace
	}

	// Set sync policy if automated sync is enabled
	if automatedSync {
		app.Spec.SyncPolicy = &SyncPolicy{
			Automated: &Automated{
				Prune:    prune,
				SelfHeal: selfHeal,
			},
		}
	}

	// Build query parameters
	queryParams := make(map[string]string)
	if validate {
		queryParams["validate"] = "true"
	}
	if upsert {
		queryParams["upsert"] = "true"
	}

	// Send request
	resp, err := c.doRequest(ctx, http.MethodPost, ArgoCDApplicationsPath, queryParams, app)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create application failed with status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var createdApp Application
	err = json.NewDecoder(resp.Body).Decode(&createdApp)
	if err != nil {
		return nil, fmt.Errorf("error parsing created application: %w", err)
	}

	return &createdApp, nil
}

// UpdateApplication updates an existing ArgoCD application
func (c *ArgoClient) UpdateApplication(ctx context.Context, name, project, repoURL, path, destServer, destNamespace, revision string,
	automatedSync, prune, selfHeal *bool, validate bool) (*Application, error) {

	// First get the current application to modify
	existingApp, err := c.GetApplication(ctx, name, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get existing application: %w", err)
	}

	// Update project if provided
	if project != "" {
		existingApp.Spec.Project = project
	}

	// Update source fields if provided
	if repoURL != "" {
		existingApp.Spec.Source.RepoURL = repoURL
	}
	if path != "" {
		existingApp.Spec.Source.Path = path
	}
	if revision != "" {
		existingApp.Spec.Source.TargetRevision = revision
	}

	// Update destination fields if provided
	if destServer != "" {
		existingApp.Spec.Destination.Server = destServer
	}
	if destNamespace != "" {
		existingApp.Spec.Destination.Namespace = destNamespace
	}

	// Update sync policy if needed
	if automatedSync != nil || prune != nil || selfHeal != nil {
		// Create sync policy if it doesn't exist
		if existingApp.Spec.SyncPolicy == nil {
			existingApp.Spec.SyncPolicy = &SyncPolicy{}
		}

		// Handle automated sync
		if automatedSync != nil {
			if *automatedSync {
				// Create automated section if it doesn't exist
				if existingApp.Spec.SyncPolicy.Automated == nil {
					existingApp.Spec.SyncPolicy.Automated = &Automated{}
				}

				// Update prune and selfHeal if provided
				if prune != nil {
					existingApp.Spec.SyncPolicy.Automated.Prune = *prune
				}
				if selfHeal != nil {
					existingApp.Spec.SyncPolicy.Automated.SelfHeal = *selfHeal
				}
			} else {
				// Remove automated section
				existingApp.Spec.SyncPolicy.Automated = nil
			}
		}
	}

	// Build query parameters
	queryParams := make(map[string]string)
	if validate {
		queryParams["validate"] = "true"
	}

	// Send update request
	updatePath := fmt.Sprintf(ArgoCDApplicationPath, name)
	resp, err := c.doRequest(ctx, http.MethodPut, updatePath, queryParams, existingApp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update application failed with status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var updatedApp Application
	err = json.NewDecoder(resp.Body).Decode(&updatedApp)
	if err != nil {
		return nil, fmt.Errorf("error parsing updated application: %w", err)
	}

	return &updatedApp, nil
}

// DeleteApplicationWithBody deletes an ArgoCD application using a body in the DELETE request
// Some ArgoCD versions or configurations expect a body with the DELETE request
func (c *ArgoClient) DeleteApplicationWithBody(ctx context.Context, name string, cascade bool, propagationPolicy, namespace string) error {
	// Create a request body - some ArgoCD versions expect this
	type DeleteRequest struct {
		Name              string `json:"name"`
		Cascade           bool   `json:"cascade"`
		PropagationPolicy string `json:"propagationPolicy,omitempty"`
		AppNamespace      string `json:"appNamespace,omitempty"`
	}

	requestBody := DeleteRequest{
		Name:    name,
		Cascade: cascade,
	}

	if propagationPolicy != "" {
		requestBody.PropagationPolicy = propagationPolicy
	}

	if namespace != "" {
		requestBody.AppNamespace = namespace
	}

	// Send delete request with body
	deletePath := fmt.Sprintf(ArgoCDApplicationPath, name)
	resp, err := c.doRequest(ctx, http.MethodDelete, deletePath, nil, requestBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete application failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteApplication deletes an ArgoCD application
// This method tries both with and without a body to accommodate different ArgoCD versions
func (c *ArgoClient) DeleteApplication(ctx context.Context, name string, cascade bool, propagationPolicy, namespace string) error {
	// First try with query parameters
	queryParams := make(map[string]string)
	queryParams["cascade"] = fmt.Sprintf("%t", cascade)

	if propagationPolicy != "" {
		queryParams["propagationPolicy"] = propagationPolicy
	}

	if namespace != "" {
		queryParams["appNamespace"] = namespace
	}

	// Send delete request
	deletePath := fmt.Sprintf(ArgoCDApplicationPath, name)

	// The ArgoCD API expects a proper DELETE request with no content body
	// but with the correct headers for authorization and optional query parameters
	resp, err := c.doRequest(ctx, http.MethodDelete, deletePath, queryParams, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// If successful, return
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return nil
	}

	// If we get a 415 error, try the alternative method with body
	if resp.StatusCode == http.StatusUnsupportedMediaType {
		return c.DeleteApplicationWithBody(ctx, name, cascade, propagationPolicy, namespace)
	}

	// Other error
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("delete application failed with status code %d: %s", resp.StatusCode, string(body))
}

// ApplicationResourceTree represents the resource tree of an application
type ApplicationResourceTree struct {
	Nodes []ResourceNode `json:"nodes"`
	Edges []ResourceEdge `json:"edges"`
}

// ResourceNode represents a node in the application resource tree
type ResourceNode struct {
	Group           string            `json:"group"`
	Version         string            `json:"version"`
	Kind            string            `json:"kind"`
	Namespace       string            `json:"namespace"`
	Name            string            `json:"name"`
	UID             string            `json:"uid"`
	ResourceVersion string            `json:"resourceVersion"`
	Health          HealthStatus      `json:"health,omitempty"`
	Status          string            `json:"status,omitempty"`
	Info            []ResourceInfo    `json:"info,omitempty"`
	NetworkingInfo  *NetworkingInfo   `json:"networkingInfo,omitempty"`
	ResourceStatus  *ResourceStatus   `json:"resourceStatus,omitempty"`
	Images          []string          `json:"images,omitempty"`
	CreatedAt       string            `json:"createdAt,omitempty"`
	ParentRefs      []ParentRef       `json:"parentRefs,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// ResourceEdge represents an edge in the application resource tree
type ResourceEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ResourceInfo represents information about a resource
type ResourceInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NetworkingInfo contains networking information about a resource
type NetworkingInfo struct {
	TargetLabels map[string]string `json:"targetLabels,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Ingress      []Ingress         `json:"ingress,omitempty"`
	ExternalURLs []string          `json:"externalURLs,omitempty"`
}

// Ingress contains information about an ingress resource
type Ingress struct {
	Host  string        `json:"host"`
	Paths []IngressPath `json:"paths"`
}

// IngressPath represents a path in an ingress
type IngressPath struct {
	Path     string `json:"path"`
	Backend  string `json:"backend"`
	Service  string `json:"service"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
}

// ParentRef represents a reference to a parent resource
type ParentRef struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	UID       string `json:"uid"`
}

// ApplicationManagedResourcesResponse represents managed resources response
type ApplicationManagedResourcesResponse struct {
	Items []ResourceDiff `json:"items"`
}

// ResourceDiff represents a diff of a resource
type ResourceDiff struct {
	Group               string `json:"group"`
	Kind                string `json:"kind"`
	Name                string `json:"name"`
	Namespace           string `json:"namespace"`
	Hook                bool   `json:"hook"`
	NormalizedLiveState string `json:"normalizedLiveState,omitempty"`
	PredictedLiveState  string `json:"predictedLiveState,omitempty"`
	TargetState         string `json:"targetState,omitempty"`
	LiveState           string `json:"liveState,omitempty"`
	Modified            bool   `json:"modified,omitempty"`
}

// ApplicationLog represents a log entry from an application
type ApplicationLog struct {
	Content   string `json:"content"`
	PodName   string `json:"podName"`
	TimeStamp string `json:"timeStamp,omitempty"`
	Container string `json:"container,omitempty"`
}

// Event represents a Kubernetes event
type Event struct {
	Metadata       Metadata    `json:"metadata"`
	Type           string      `json:"type"`
	Reason         string      `json:"reason"`
	Message        string      `json:"message"`
	Count          int         `json:"count"`
	FirstTimestamp string      `json:"firstTimestamp"`
	LastTimestamp  string      `json:"lastTimestamp"`
	InvolvedObject ObjectRef   `json:"involvedObject"`
	Source         EventSource `json:"source"`
}

// EventSource represents the source of an event
type EventSource struct {
	Component string `json:"component,omitempty"`
	Host      string `json:"host,omitempty"`
}

// ObjectRef contains reference to an object
type ObjectRef struct {
	Kind            string `json:"kind"`
	Namespace       string `json:"namespace,omitempty"`
	Name            string `json:"name"`
	UID             string `json:"uid,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// EventList represents a list of events
type EventList struct {
	Items []Event `json:"items"`
}

// ResourceAction represents an action that can be performed on a resource
type ResourceAction struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled"`
	Background  bool   `json:"background"`
}

// ResourceActionsResponse represents available actions for a resource
type ResourceActionsResponse struct {
	Actions []ResourceAction `json:"actions"`
}

// GetApplicationResourceTree gets the resource tree for an application
func (c *ArgoClient) GetApplicationResourceTree(ctx context.Context, appName string) (*ApplicationResourceTree, error) {
	path := fmt.Sprintf("/api/v1/applications/%s/resource-tree", appName)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get application resource tree failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var resourceTree ApplicationResourceTree
	err = json.NewDecoder(resp.Body).Decode(&resourceTree)
	if err != nil {
		return nil, fmt.Errorf("error parsing application resource tree: %w", err)
	}

	return &resourceTree, nil
}

// GetApplicationManagedResources gets the managed resources for an application
func (c *ArgoClient) GetApplicationManagedResources(ctx context.Context, appName string) (*ApplicationManagedResourcesResponse, error) {
	path := fmt.Sprintf("/api/v1/applications/%s/managed-resources", appName)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get application managed resources failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var managedResources ApplicationManagedResourcesResponse
	err = json.NewDecoder(resp.Body).Decode(&managedResources)
	if err != nil {
		return nil, fmt.Errorf("error parsing application managed resources: %w", err)
	}

	return &managedResources, nil
}

// GetWorkloadLogs gets logs for a workload in an application
func (c *ArgoClient) GetWorkloadLogs(ctx context.Context, appName, appNamespace string, resourceRef ResourceRef, follow bool, tailLines string) ([]ApplicationLog, error) {
	path := fmt.Sprintf("/api/v1/applications/%s/logs", appName)

	// Build query parameters
	queryParams := make(map[string]string)
	queryParams["appNamespace"] = appNamespace
	queryParams["namespace"] = resourceRef.Namespace
	queryParams["name"] = resourceRef.Name
	queryParams["kind"] = resourceRef.Kind
	queryParams["group"] = resourceRef.Group
	queryParams["resourceVersion"] = resourceRef.Version

	if resourceRef.Container != "" {
		queryParams["container"] = resourceRef.Container
	}

	// Use provided tail lines or default
	if tailLines != "" {
		queryParams["tailLines"] = tailLines
	} else {
		queryParams["tailLines"] = "100" // Default to 100 lines
	}

	// Set follow parameter
	if follow {
		queryParams["follow"] = "true"
	} else {
		queryParams["follow"] = "false"
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get workload logs failed with status code %d: %s", resp.StatusCode, string(body))
	}

	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading logs response: %w", err)
	}

	// Try to parse as JSON first
	var logs []ApplicationLog
	if err := json.Unmarshal(bodyBytes, &logs); err != nil {
		// If JSON parsing fails, treat as plain text
		logLines := strings.Split(string(bodyBytes), "\n")
		for _, line := range logLines {
			if line != "" {
				logs = append(logs, ApplicationLog{
					Content:   line,
					PodName:   resourceRef.Name,
					Container: resourceRef.Container,
				})
			}
		}
	}

	return logs, nil
}

// GetResourceEvents gets events for a resource in an application
func (c *ArgoClient) GetResourceEvents(ctx context.Context, appName, appNamespace, resourceNamespace, resourceName string) (*EventList, error) {
	path := fmt.Sprintf("/api/v1/applications/%s/events", appName)

	queryParams := make(map[string]string)
	queryParams["appNamespace"] = appNamespace
	queryParams["resourceNamespace"] = resourceNamespace
	queryParams["resourceName"] = resourceName

	resp, err := c.doRequest(ctx, http.MethodGet, path, queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var events EventList
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to parse events response: %w", err)
	}

	return &events, nil
}

// GetApplicationEvents gets events for an application
func (c *ArgoClient) GetApplicationEvents(ctx context.Context, appName, appNamespace string) (*EventList, error) {
	path := fmt.Sprintf("/api/v1/applications/%s/events", appName)

	queryParams := make(map[string]string)
	if appNamespace != "" {
		queryParams["appNamespace"] = appNamespace
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var events EventList
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to parse events response: %w", err)
	}

	return &events, nil
}

// GetResourceActions gets available actions for a resource in an application
func (c *ArgoClient) GetResourceActions(ctx context.Context, appName, appNamespace, resourceNamespace, resourceName,
	resourceKind, resourceGroup, resourceVersion string) (*ResourceActionsResponse, error) {

	path := fmt.Sprintf("/api/v1/applications/%s/resource/actions", appName)

	// Build query parameters
	queryParams := make(map[string]string)
	queryParams["appNamespace"] = appNamespace
	queryParams["namespace"] = resourceNamespace
	queryParams["name"] = resourceName
	queryParams["kind"] = resourceKind

	if resourceGroup != "" {
		queryParams["group"] = resourceGroup
	}

	if resourceVersion != "" {
		queryParams["version"] = resourceVersion
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get resource actions failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var actions ResourceActionsResponse
	err = json.NewDecoder(resp.Body).Decode(&actions)
	if err != nil {
		return nil, fmt.Errorf("error parsing resource actions: %w", err)
	}

	return &actions, nil
}

// RunResourceAction runs an action on a resource in an application
func (c *ArgoClient) RunResourceAction(ctx context.Context, appName, appNamespace, resourceNamespace, resourceName,
	resourceKind, resourceGroup, resourceVersion, actionName string) (*Application, error) {

	path := fmt.Sprintf("/api/v1/applications/%s/resource/actions", appName)

	// Build query parameters
	queryParams := make(map[string]string)
	queryParams["appNamespace"] = appNamespace
	queryParams["namespace"] = resourceNamespace
	queryParams["name"] = resourceName
	queryParams["kind"] = resourceKind

	if resourceGroup != "" {
		queryParams["group"] = resourceGroup
	}

	if resourceVersion != "" {
		queryParams["version"] = resourceVersion
	}

	// Action name is passed in the body
	actionBody := map[string]string{
		"action": actionName,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, path, queryParams, actionBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("run resource action failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var result Application
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error parsing action result: %w", err)
	}

	return &result, nil
}
