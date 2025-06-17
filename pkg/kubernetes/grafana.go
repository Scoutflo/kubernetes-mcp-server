package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// getGrafanaURL gets the Grafana URL from environment variable
func (k *Kubernetes) getGrafanaURL() (string, error) {
	grafanaURL := os.Getenv("GRAFANA_URL")
	if grafanaURL == "" {
		return "", fmt.Errorf("GRAFANA_URL environment variable is not set")
	}

	// Ensure the URL doesn't end with a slash for consistent path construction
	grafanaURL = strings.TrimSuffix(grafanaURL, "/")

	return grafanaURL, nil
}

// getGrafanaAPIKey gets the Grafana API key from environment variable
func (k *Kubernetes) getGrafanaAPIKey() (string, error) {
	apiKey := os.Getenv("GRAFANA_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GRAFANA_API_KEY environment variable is not set")
	}
	return apiKey, nil
}

// makeGrafanaRequest creates and executes an HTTP request to Grafana API
func (k *Kubernetes) makeGrafanaRequest(endpoint string) (map[string]interface{}, error) {
	return k.makeGrafanaRequestWithMethod("GET", endpoint, nil)
}

// makeGrafanaRequestWithMethod creates and executes an HTTP request to Grafana API with specified method and body
func (k *Kubernetes) makeGrafanaRequestWithMethod(method, endpoint string, body []byte) (map[string]interface{}, error) {
	// Get Grafana URL and API Key
	grafanaURL, err := k.getGrafanaURL()
	if err != nil {
		return nil, err
	}

	apiKey, err := k.getGrafanaAPIKey()
	if err != nil {
		return nil, err
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("%s%s", grafanaURL, endpoint)

	// Create HTTP request
	var req *http.Request
	if body != nil {
		req, err = http.NewRequest(method, apiURL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, apiURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Grafana: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Try to read response body for better error reporting
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil && len(bodyBytes) > 0 {
			return nil, fmt.Errorf("Grafana API returned status %d: %s. Response: %s", resp.StatusCode, resp.Status, string(bodyBytes))
		}
		return nil, fmt.Errorf("Grafana API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Grafana response: %v", err)
	}

	return result, nil
}

// GetDashboardByUID retrieves a dashboard by its UID from Grafana
func (k *Kubernetes) GetDashboardByUID(uid string) (string, error) {
	endpoint := fmt.Sprintf("/api/dashboards/uid/%s", uid)

	result, err := k.makeGrafanaRequest(endpoint)
	if err != nil {
		return "", err
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Grafana result: %v", err)
	}

	return string(resultJSON), nil
}

// SearchDashboards searches for dashboards in Grafana by query string
func (k *Kubernetes) SearchDashboards(query string) (string, error) {
	// Get Grafana URL and API Key
	grafanaURL, err := k.getGrafanaURL()
	if err != nil {
		return "", err
	}

	apiKey, err := k.getGrafanaAPIKey()
	if err != nil {
		return "", err
	}

	// Build the search endpoint with query parameter and type filter for dashboards
	endpoint := "/api/search?type=dash-db"
	if query != "" {
		endpoint = fmt.Sprintf("/api/search?query=%s&type=dash-db", url.QueryEscape(query))
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("%s%s", grafanaURL, endpoint)

	// Create HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Grafana: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Try to read response body for better error reporting
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil && len(bodyBytes) > 0 {
			return "", fmt.Errorf("Grafana API returned status %d: %s. Response: %s", resp.StatusCode, resp.Status, string(bodyBytes))
		}
		return "", fmt.Errorf("Grafana API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse the response as an array (search API returns array of dashboards)
	var dashboards []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&dashboards); err != nil {
		return "", fmt.Errorf("failed to decode Grafana response: %v", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(dashboards)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Grafana result: %v", err)
	}

	return string(resultJSON), nil
}

// UpdateDashboard creates or updates a dashboard in Grafana
func (k *Kubernetes) UpdateDashboard(dashboard map[string]interface{}, folderUID, message string, overwrite bool, userID int64) (string, error) {
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

	// Make POST request to update/create dashboard
	endpoint := "/api/dashboards/db"
	result, err := k.makeGrafanaRequestWithMethod("POST", endpoint, payloadJSON)
	if err != nil {
		return "", err
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Grafana result: %v", err)
	}

	return string(resultJSON), nil
}

// GetDashboardPanelQueries extracts panel queries from a dashboard by UID
func (k *Kubernetes) GetDashboardPanelQueries(uid string) (string, error) {
	// First, get the dashboard
	dashboardJSON, err := k.GetDashboardByUID(uid)
	if err != nil {
		return "", fmt.Errorf("failed to get dashboard: %v", err)
	}

	// Parse the dashboard JSON
	var dashboardResponse map[string]interface{}
	if err := json.Unmarshal([]byte(dashboardJSON), &dashboardResponse); err != nil {
		return "", fmt.Errorf("failed to parse dashboard JSON: %v", err)
	}

	// Extract the dashboard object
	dashboard, ok := dashboardResponse["dashboard"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("dashboard field is not a JSON object")
	}

	// Extract panels array
	panelsInterface, ok := dashboard["panels"]
	if !ok {
		return "", fmt.Errorf("dashboard does not contain panels")
	}

	panels, ok := panelsInterface.([]interface{})
	if !ok {
		return "", fmt.Errorf("panels is not a JSON array")
	}

	// Structure for panel query result
	type datasourceInfo struct {
		UID  string `json:"uid"`
		Type string `json:"type"`
	}

	type panelQuery struct {
		Title      string         `json:"title"`
		Query      string         `json:"query"`
		Datasource datasourceInfo `json:"datasource"`
	}

	result := make([]panelQuery, 0)

	// Process each panel
	for _, p := range panels {
		panel, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract panel title
		title, _ := panel["title"].(string)

		// Extract datasource information
		var dsInfo datasourceInfo
		if dsField, dsExists := panel["datasource"]; dsExists && dsField != nil {
			if dsMap, ok := dsField.(map[string]interface{}); ok {
				if uid, ok := dsMap["uid"].(string); ok {
					dsInfo.UID = uid
				}
				if dsType, ok := dsMap["type"].(string); ok {
					dsInfo.Type = dsType
				}
			}
		}

		// Extract targets (queries) from the panel
		targetsInterface, ok := panel["targets"]
		if !ok {
			continue
		}

		targets, ok := targetsInterface.([]interface{})
		if !ok {
			continue
		}

		// Process each target in the panel
		for _, t := range targets {
			target, ok := t.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract query expression - try different field names as different datasources use different field names
			var query string
			if expr, ok := target["expr"].(string); ok && expr != "" {
				query = expr // Prometheus queries
			} else if rawSql, ok := target["rawSql"].(string); ok && rawSql != "" {
				query = rawSql // SQL queries
			} else if queryStr, ok := target["query"].(string); ok && queryStr != "" {
				query = queryStr // Generic query field
			}

			// Only add if we found a query
			if query != "" {
				result = append(result, panelQuery{
					Title:      title,
					Query:      query,
					Datasource: dsInfo,
				})
			}
		}
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal panel queries result: %v", err)
	}

	return string(resultJSON), nil
}

// ListDatasources lists all available datasources, optionally filtered by type
func (k *Kubernetes) ListDatasources(dsType string) (string, error) {
	// Get Grafana URL and API Key
	grafanaURL, err := k.getGrafanaURL()
	if err != nil {
		return "", err
	}

	apiKey, err := k.getGrafanaAPIKey()
	if err != nil {
		return "", err
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("%s/api/datasources", grafanaURL)

	// Create HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Grafana: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Grafana API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse the response as an array
	var datasources []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&datasources); err != nil {
		return "", fmt.Errorf("failed to decode Grafana response: %v", err)
	}

	// If type filter is specified, filter the datasources
	if dsType != "" {
		// Filter datasources by type (case-insensitive contains match)
		filteredDatasources := make([]interface{}, 0)
		for _, ds := range datasources {
			if dsMap, ok := ds.(map[string]interface{}); ok {
				if dsTypeField, exists := dsMap["type"]; exists {
					if dsTypeStr, ok := dsTypeField.(string); ok {
						// Case-insensitive contains match
						if strings.Contains(strings.ToLower(dsTypeStr), strings.ToLower(dsType)) {
							filteredDatasources = append(filteredDatasources, ds)
						}
					}
				}
			}
		}
		datasources = filteredDatasources
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(datasources)
	if err != nil {
		return "", fmt.Errorf("failed to marshal datasources result: %v", err)
	}

	return string(resultJSON), nil
}

// GetDatasourceByUID retrieves a datasource by its UID from Grafana
func (k *Kubernetes) GetDatasourceByUID(uid string) (string, error) {
	endpoint := fmt.Sprintf("/api/datasources/uid/%s", uid)

	result, err := k.makeGrafanaRequest(endpoint)
	if err != nil {
		// Check if it's a 404 error and provide a better message
		if strings.Contains(err.Error(), "404") {
			return "", fmt.Errorf("datasource with UID '%s' not found. Please check if the datasource exists and is accessible", uid)
		}
		return "", err
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal datasource result: %v", err)
	}

	return string(resultJSON), nil
}

// GetDatasourceByName retrieves a datasource by its name from Grafana
func (k *Kubernetes) GetDatasourceByName(name string) (string, error) {
	endpoint := fmt.Sprintf("/api/datasources/name/%s", url.QueryEscape(name))

	result, err := k.makeGrafanaRequest(endpoint)
	if err != nil {
		return "", err
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal datasource result: %v", err)
	}

	return string(resultJSON), nil
}

// HealthCheck tests basic connectivity to Grafana
func (k *Kubernetes) HealthCheck() (string, error) {
	// Get credentials for debugging (don't log the actual key)
	grafanaURL, err := k.getGrafanaURL()
	if err != nil {
		return "", fmt.Errorf("failed to get Grafana URL: %v", err)
	}

	apiKey, err := k.getGrafanaAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to get Grafana API key: %v", err)
	}

	// Create debug info (without exposing sensitive data)
	debugInfo := map[string]interface{}{
		"grafana_url":    grafanaURL,
		"api_key_set":    apiKey != "",
		"api_key_length": len(apiKey),
	}

	// Try the org endpoint first (simpler than health)
	endpoint := "/api/org"

	result, err := k.makeGrafanaRequest(endpoint)
	if err != nil {
		// Include debug info in error
		debugJSON, _ := json.Marshal(debugInfo)
		return "", fmt.Errorf("Grafana health check failed.\nDebug info:\n%s\nError: %v", string(debugJSON), err)
	}

	// Add debug info to successful result
	result["debug_info"] = debugInfo

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal health check result: %v", err)
	}

	return string(resultJSON), nil
}

// ListAlertRules lists Grafana alert rules with optional filtering and pagination
func (k *Kubernetes) ListAlertRules(limit, page int, labelSelectors []map[string]interface{}) (string, error) {
	endpoint := "/api/ruler/grafana/api/v1/rules"

	result, err := k.makeGrafanaRequest(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to get alert rules: %v", err)
	}

	// Process the alert rules and apply filtering/pagination
	alertRules, err := k.processAlertRules(result, limit, page, labelSelectors)
	if err != nil {
		return "", fmt.Errorf("failed to process alert rules: %v", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(alertRules)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alert rules result: %v", err)
	}

	return string(resultJSON), nil
}

// processAlertRules processes raw alert rules data and applies filtering/pagination
func (k *Kubernetes) processAlertRules(rawData map[string]interface{}, limit, page int, labelSelectors []map[string]interface{}) ([]map[string]interface{}, error) {
	var allRules []map[string]interface{}

	// Navigate through the nested structure to extract rules
	if groups, ok := rawData["groups"].([]interface{}); ok {
		for _, group := range groups {
			if groupMap, ok := group.(map[string]interface{}); ok {
				if rules, ok := groupMap["rules"].([]interface{}); ok {
					for _, rule := range rules {
						if ruleMap, ok := rule.(map[string]interface{}); ok {
							// Create a summary of the rule
							ruleSummary := map[string]interface{}{
								"uid":    ruleMap["uid"],
								"title":  ruleMap["title"],
								"state":  ruleMap["state"],
								"labels": ruleMap["labels"],
							}

							// Apply label selector filtering if provided
							if len(labelSelectors) == 0 || k.matchesLabelSelectors(ruleMap, labelSelectors) {
								allRules = append(allRules, ruleSummary)
							}
						}
					}
				}
			}
		}
	}

	// Apply pagination
	if limit <= 0 {
		limit = 100 // default
	}
	if page <= 0 {
		page = 1 // default
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= len(allRules) {
		return []map[string]interface{}{}, nil
	}

	if end > len(allRules) {
		end = len(allRules)
	}

	return allRules[start:end], nil
}

// matchesLabelSelectors checks if a rule matches the provided label selectors
func (k *Kubernetes) matchesLabelSelectors(rule map[string]interface{}, selectors []map[string]interface{}) bool {
	labels, ok := rule["labels"].(map[string]interface{})
	if !ok {
		return len(selectors) == 0
	}

	for _, selector := range selectors {
		name, nameOk := selector["name"].(string)
		selectorType, typeOk := selector["type"].(string)
		value, valueOk := selector["value"].(string)

		if !nameOk || !typeOk || !valueOk {
			continue // Skip invalid selectors
		}

		labelValue, labelExists := labels[name].(string)

		switch selectorType {
		case "=":
			if !labelExists || labelValue != value {
				return false
			}
		case "!=":
			if labelExists && labelValue == value {
				return false
			}
		}
	}

	return true
}

// GetAlertRuleByUID retrieves a specific alert rule by its UID
func (k *Kubernetes) GetAlertRuleByUID(uid string) (string, error) {
	endpoint := fmt.Sprintf("/api/v1/provisioning/alert-rules/%s", uid)

	result, err := k.makeGrafanaRequest(endpoint)
	if err != nil {
		// Check if it's a 404 error and provide a better message
		if strings.Contains(err.Error(), "404") {
			return "", fmt.Errorf("alert rule with UID '%s' not found. Please check if the alert rule exists and is accessible", uid)
		}
		return "", fmt.Errorf("failed to get alert rule: %v", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alert rule result: %v", err)
	}

	return string(resultJSON), nil
}
