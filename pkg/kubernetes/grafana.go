package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GetDashboardByUID retrieves a dashboard by its UID from Grafana
func (k *Kubernetes) GetDashboardByUID(ctx context.Context, uid string) (string, error) {
	if uid == "" {
		return "", fmt.Errorf("uid parameter is required")
	}

	endpoint := fmt.Sprintf("/apis/v1/grafana/dashboard-by-uid?uid=%s", url.QueryEscape(uid))
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get dashboard by UID: %w", err)
	}
	return string(response), nil
}

// SearchDashboards searches for dashboards in Grafana by query string
func (k *Kubernetes) SearchDashboards(ctx context.Context, query string) (string, error) {
	endpoint := "/apis/v1/grafana/search-dashboards"
	if query != "" {
		endpoint = fmt.Sprintf("/apis/v1/grafana/search-dashboards?query=%s", url.QueryEscape(query))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to search dashboards: %w", err)
	}
	return string(response), nil
}

// UpdateDashboard creates or updates a dashboard in Grafana
func (k *Kubernetes) UpdateDashboard(ctx context.Context, dashboard map[string]interface{}, folderUID, message string, overwrite bool, userID int64) (string, error) {
	// Build the request payload
	payload := map[string]interface{}{
		"dashboard": dashboard,
		"overwrite": overwrite,
	}

	if folderUID != "" {
		payload["folderUid"] = folderUID
	}

	if message != "" {
		payload["message"] = message
	}

	if userID != 0 {
		payload["userId"] = userID
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal dashboard payload: %v", err)
	}

	endpoint := "/apis/v1/grafana/update-dashboard"
	response, err := k.MakeAPIRequest("POST", endpoint, payloadJSON)
	if err != nil {
		return "", fmt.Errorf("failed to update dashboard: %w", err)
	}
	return string(response), nil
}

// GetDashboardPanelQueries extracts panel queries from a dashboard by UID
func (k *Kubernetes) GetDashboardPanelQueries(ctx context.Context, uid string) (string, error) {
	if uid == "" {
		return "", fmt.Errorf("uid parameter is required")
	}

	endpoint := fmt.Sprintf("/apis/v1/grafana/dashboard-panel-queries?uid=%s", url.QueryEscape(uid))
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get dashboard panel queries: %w", err)
	}
	return string(response), nil
}

// ListDatasources lists all available datasources, optionally filtered by type
func (k *Kubernetes) ListDatasources(ctx context.Context, dsType string) (string, error) {
	endpoint := "/apis/v1/grafana/datasources"
	if dsType != "" {
		endpoint = fmt.Sprintf("/apis/v1/grafana/datasources?type=%s", url.QueryEscape(dsType))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list datasources: %w", err)
	}
	return string(response), nil
}

// GetDatasourceByUID retrieves a datasource by its UID from Grafana
func (k *Kubernetes) GetDatasourceByUID(ctx context.Context, uid string) (string, error) {
	if uid == "" {
		return "", fmt.Errorf("uid parameter is required")
	}

	endpoint := fmt.Sprintf("/apis/v1/grafana/datasource-by-uid?uid=%s", url.QueryEscape(uid))
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get datasource by UID: %w", err)
	}
	return string(response), nil
}

// GetDatasourceByName retrieves a datasource by its name from Grafana
func (k *Kubernetes) GetDatasourceByName(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name parameter is required")
	}

	endpoint := fmt.Sprintf("/apis/v1/grafana/datasource-by-name?name=%s", url.QueryEscape(name))
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get datasource by name: %w", err)
	}
	return string(response), nil
}

// HealthCheck tests basic connectivity to Grafana
func (k *Kubernetes) HealthCheck(ctx context.Context) (string, error) {
	endpoint := "/apis/v1/grafana/health"
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to perform Grafana health check: %w", err)
	}
	return string(response), nil
}

// ListAlertRules lists Grafana alert rules with optional filtering and pagination
func (k *Kubernetes) ListAlertRules(ctx context.Context, limit, page int, labelSelectors []map[string]interface{}) (string, error) {
	endpoint := "/apis/v1/grafana/alert-rules"

	var params []string

	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}

	if page > 0 {
		params = append(params, fmt.Sprintf("page=%d", page))
	}

	if len(labelSelectors) > 0 {
		selectorsJSON, err := json.Marshal(labelSelectors)
		if err != nil {
			return "", fmt.Errorf("failed to marshal label selectors: %v", err)
		}
		params = append(params, fmt.Sprintf("label_selectors=%s", url.QueryEscape(string(selectorsJSON))))
	}

	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list alert rules: %w", err)
	}
	return string(response), nil
}

// GetAlertRuleByUID retrieves a specific alert rule by its UID
func (k *Kubernetes) GetAlertRuleByUID(ctx context.Context, uid string) (string, error) {
	if uid == "" {
		return "", fmt.Errorf("uid parameter is required")
	}

	endpoint := fmt.Sprintf("/apis/v1/grafana/alert-rule-by-uid?uid=%s", url.QueryEscape(uid))
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get alert rule by UID: %w", err)
	}
	return string(response), nil
}
