package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scoutflo/kubernetes-mcp-server/pkg/llm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getPrometheusURL attempts to discover the Prometheus URL in the cluster
func (k *Kubernetes) getPrometheusURL() (string, error) {
	// Try to find Prometheus Ingress first (preferred method)
	namespace := "scoutflo-monitoring"

	// Get Ingress resources in the monitoring namespace
	networkingV1 := k.clientSet.NetworkingV1()
	ingresses, err := networkingV1.Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil && len(ingresses.Items) > 0 {
		// Look for Ingress that routes to Prometheus
		for _, ingress := range ingresses.Items {
			// Check all rules
			for _, rule := range ingress.Spec.Rules {
				// Check if the host looks like a monitoring host
				if strings.Contains(rule.Host, "scoutflo-monitoring") ||
					strings.Contains(rule.Host, "prometheus") {

					// Check all paths
					if rule.HTTP != nil {
						for _, path := range rule.HTTP.Paths {
							if strings.Contains(path.Backend.Service.Name, "prometheus") &&
								!strings.Contains(path.Backend.Service.Name, "alertmanager") {

								// Found a matching Ingress rule for Prometheus
								// Check if TLS is configured
								protocol := "http"
								for _, tls := range ingress.Spec.TLS {
									for _, host := range tls.Hosts {
										if host == rule.Host || host == "*.scoutflo.agency" {
											protocol = "https"
											break
										}
									}
								}

								// Construct URL with path if specified
								pathPrefix := "/"
								if path.Path != "" {
									pathPrefix = path.Path
								}

								prometheusURL := fmt.Sprintf("%s://%s%s", protocol, rule.Host, pathPrefix)
								return prometheusURL, nil
							}
						}
					}
				}
			}
		}
	}

	// If we couldn't find Prometheus via Ingress, try to port-forward the service
	// Note: This is a fallback and may need adaptation based on your environment
	return "", fmt.Errorf("prometheus URL not found via ingress")
}

// QueryPrometheus sends an instant query to Prometheus
func (k *Kubernetes) QueryPrometheus(query string, queryTime *time.Time, timeout string) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API query URL with additional parameters
	baseURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, url.QueryEscape(query))

	// Add timestamp if provided
	if queryTime != nil && !queryTime.IsZero() {
		baseURL = fmt.Sprintf("%s&time=%d", baseURL, queryTime.Unix())
	}

	// Add timeout if provided
	if timeout != "" {
		baseURL = fmt.Sprintf("%s&timeout=%s", baseURL, url.QueryEscape(timeout))
	}

	// Send the request
	resp, err := http.Get(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check for error status
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorType, ok := result["errorType"].(string); ok {
			if errorMsg, ok := result["error"].(string); ok {
				return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
			}
			return "", fmt.Errorf("Prometheus error: %s", errorType)
		}
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Check for no data
	if data, ok := result["data"].(map[string]interface{}); ok {
		if resultArr, ok := data["result"].([]interface{}); ok && len(resultArr) == 0 {
			// Return a specific message for no data rather than an error
			emptyResult := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "vector",
					"result":     []interface{}{},
				},
				"info": "No data found for the query. This could be due to an incorrect metric name or no data points available in the specified time range.",
			}

			resultJSON, err := json.MarshalIndent(emptyResult, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal empty result: %v", err)
			}
			return string(resultJSON), nil
		}
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// QueryPrometheusRange sends a range query to Prometheus
func (k *Kubernetes) QueryPrometheusRange(query string, start, end time.Time, step string, timeout string) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Convert times to Unix timestamps
	startTS := start.Unix()
	endTS := end.Unix()

	// Build the API query URL for range queries
	baseURL := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%s",
		prometheusURL, url.QueryEscape(query), startTS, endTS, url.QueryEscape(step))

	// Add timeout if provided
	if timeout != "" {
		baseURL = fmt.Sprintf("%s&timeout=%s", baseURL, url.QueryEscape(timeout))
	}

	// Send the request
	resp, err := http.Get(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check for error status
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorType, ok := result["errorType"].(string); ok {
			if errorMsg, ok := result["error"].(string); ok {
				return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
			}
			return "", fmt.Errorf("Prometheus error: %s", errorType)
		}
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Check for no data
	if data, ok := result["data"].(map[string]interface{}); ok {
		if resultArr, ok := data["result"].([]interface{}); ok && len(resultArr) == 0 {
			// Return a specific message for no data rather than an error
			emptyResult := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "matrix",
					"result":     []interface{}{},
				},
				"info": "No data found for the query. This could be due to an incorrect metric name or no data points available in the specified time range.",
			}

			resultJSON, err := json.MarshalIndent(emptyResult, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal empty result: %v", err)
			}
			return string(resultJSON), nil
		}
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// ListPrometheusMetrics retrieves all available metrics from Prometheus
func (k *Kubernetes) ListPrometheusMetrics() (string, error) {
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for metadata endpoint
	apiURL := fmt.Sprintf("%s/api/v1/metadata", prometheusURL)

	// Send the request
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to list Prometheus metrics: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusMetricInfo retrieves detailed information about a specific metric
func (k *Kubernetes) GetPrometheusMetricInfo(metricName string, includeStats bool) (string, error) {
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for metadata of specific metric
	metadataURL := fmt.Sprintf("%s/api/v1/metadata?metric=%s", prometheusURL, url.QueryEscape(metricName))

	// Send the request
	metadataResp, err := http.Get(metadataURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus metric metadata: %v", err)
	}
	defer metadataResp.Body.Close()

	// Parse the metadata response
	var metadataResult map[string]interface{}
	if err := json.NewDecoder(metadataResp.Body).Decode(&metadataResult); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus metadata response: %v", err)
	}

	result := map[string]interface{}{
		"name":     metricName,
		"metadata": metadataResult,
	}

	// If statistics are requested, fetch them
	if includeStats {
		statsResult, err := k.getMetricStatistics(prometheusURL, metricName)
		if err != nil {
			return "", err
		}
		result["statistics"] = statsResult
	}

	// Convert final result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// getMetricStatistics calculates basic statistics for a metric
func (k *Kubernetes) getMetricStatistics(prometheusURL, metricName string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Define the queries for min, max, and count
	queries := map[string]string{
		"count": fmt.Sprintf("count(%s)", metricName),
		"min":   fmt.Sprintf("min(%s)", metricName),
		"max":   fmt.Sprintf("max(%s)", metricName),
		"avg":   fmt.Sprintf("avg(%s)", metricName),
	}

	// Execute each query and collect results
	for statName, query := range queries {
		queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, url.QueryEscape(query))
		resp, err := http.Get(queryURL)
		if err != nil {
			return nil, fmt.Errorf("failed to query %s statistic: %v", statName, err)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode %s statistic response: %v", statName, err)
		}
		resp.Body.Close()

		// Extract the value from Prometheus result
		if status, ok := result["status"].(string); ok && status == "success" {
			if data, ok := result["data"].(map[string]interface{}); ok {
				if resultData, ok := data["result"].([]interface{}); ok && len(resultData) > 0 {
					if resultItem, ok := resultData[0].(map[string]interface{}); ok {
						if value, ok := resultItem["value"].([]interface{}); ok && len(value) > 1 {
							stats[statName] = value[1]
						}
					}
				}
			}
		}
	}

	return stats, nil
}

// GeneratePromQLQuery generates a PromQL query from a natural language description
func (k *Kubernetes) GeneratePromQLQuery(description string) (string, error) {
	// Create a new LLM client
	llmClient, err := llm.NewDefaultClient()
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %v", err)
	}

	// Make the LLM API call with the PromQL prompt and the description
	response, err := llmClient.Call(llm.PromQLPrompt, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate PromQL query: %v", err)
	}

	return response, nil
}

// QueryPrometheusSeries queries series from Prometheus
func (k *Kubernetes) QueryPrometheusSeries(match []string, start, end *time.Time, limit int) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the base URL
	baseURL := fmt.Sprintf("%s/api/v1/series", prometheusURL)

	// Build query parameters
	queryParams := url.Values{}

	// Add match parameters (can be multiple)
	for _, m := range match {
		queryParams.Add("match[]", m)
	}

	// Add start time if provided
	if start != nil && !start.IsZero() {
		queryParams.Add("start", fmt.Sprintf("%d", start.Unix()))
	}

	// Add end time if provided
	if end != nil && !end.IsZero() {
		queryParams.Add("end", fmt.Sprintf("%d", end.Unix()))
	}

	// Add limit if provided and valid
	if limit > 0 {
		queryParams.Add("limit", fmt.Sprintf("%d", limit))
	}

	// Combine base URL with query parameters
	requestURL := fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())

	// Send the request
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus series: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check for error status
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorType, ok := result["errorType"].(string); ok {
			if errorMsg, ok := result["error"].(string); ok {
				return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
			}
			return "", fmt.Errorf("Prometheus error: %s", errorType)
		}
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Check for no data
	if data, ok := result["data"].([]interface{}); ok && len(data) == 0 {
		// Return a specific message for no data rather than an error
		emptyResult := map[string]interface{}{
			"status": "success",
			"data":   []interface{}{},
			"info":   "No series found matching the criteria.",
		}

		resultJSON, err := json.MarshalIndent(emptyResult, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal empty result: %v", err)
		}
		return string(resultJSON), nil
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusTargets retrieves information about Prometheus targets
func (k *Kubernetes) GetPrometheusTargets(state, scrapePool string) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the base URL
	baseURL := fmt.Sprintf("%s/api/v1/targets", prometheusURL)

	// Build query parameters
	queryParams := url.Values{}

	// Add state filter if provided
	if state != "" {
		// Validate state
		if state != "active" && state != "dropped" && state != "any" {
			return "", fmt.Errorf("invalid state parameter: %s (must be one of: active, dropped, any)", state)
		}
		queryParams.Add("state", state)
	}

	// Add scrape pool filter if provided
	if scrapePool != "" {
		queryParams.Add("scrape_pool", scrapePool)
	}

	// Combine base URL with query parameters
	requestURL := fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())

	// Send the request
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus targets: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check for error status
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorType, ok := result["errorType"].(string); ok {
			if errorMsg, ok := result["error"].(string); ok {
				return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
			}
			return "", fmt.Errorf("Prometheus error: %s", errorType)
		}
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusTargetMetadata retrieves metadata about metrics exposed by targets
func (k *Kubernetes) GetPrometheusTargetMetadata(matchTarget, metric string, limit int) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the base URL
	baseURL := fmt.Sprintf("%s/api/v1/targets/metadata", prometheusURL)

	// Build query parameters
	queryParams := url.Values{}

	// Add match_target parameter if provided
	if matchTarget != "" {
		queryParams.Add("match_target", matchTarget)
	}

	// Add metric parameter if provided
	if metric != "" {
		queryParams.Add("metric", metric)
	}

	// Add limit parameter if provided and valid
	if limit > 0 {
		queryParams.Add("limit", fmt.Sprintf("%d", limit))
	}

	// Combine base URL with query parameters
	requestURL := fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())

	// Send the request
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus target metadata: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check for error status
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorType, ok := result["errorType"].(string); ok {
			if errorMsg, ok := result["error"].(string); ok {
				return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
			}
			return "", fmt.Errorf("Prometheus error: %s", errorType)
		}
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Check for no data
	if data, ok := result["data"].([]interface{}); ok && len(data) == 0 {
		// Return a specific message for no data rather than an error
		emptyResult := map[string]interface{}{
			"status": "success",
			"data":   []interface{}{},
			"info":   "No metadata found matching the criteria.",
		}

		resultJSON, err := json.MarshalIndent(emptyResult, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal empty result: %v", err)
		}
		return string(resultJSON), nil
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusAlerts retrieves all currently firing alerts from Prometheus
func (k *Kubernetes) GetPrometheusAlerts() (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for alerts endpoint
	alertsURL := fmt.Sprintf("%s/api/v1/alerts", prometheusURL)

	// Send the request
	resp, err := http.Get(alertsURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus alerts: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check for error status
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorType, ok := result["errorType"].(string); ok {
			if errorMsg, ok := result["error"].(string); ok {
				return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
			}
			return "", fmt.Errorf("Prometheus error: %s", errorType)
		}
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Format the result with additional information for no alerts case
	if data, ok := result["data"].(map[string]interface{}); ok {
		if alerts, ok := data["alerts"].([]interface{}); ok && len(alerts) == 0 {
			result["info"] = "No alerts are currently firing."
		}
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusRules retrieves information about configured alerting and recording rules
func (k *Kubernetes) GetPrometheusRules(ruleType, groupLimit string, ruleNames, ruleGroups, files []string, excludeAlerts bool, matchLabels []string) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for rules endpoint
	baseURL := fmt.Sprintf("%s/api/v1/rules", prometheusURL)

	// Build query parameters
	queryParams := url.Values{}

	// Add rule type filter if provided
	if ruleType != "" {
		queryParams.Add("type", ruleType)
	}

	// Add rule names filter if provided
	for _, name := range ruleNames {
		if name != "" {
			queryParams.Add("rule_name[]", name)
		}
	}

	// Add rule group filter if provided
	for _, group := range ruleGroups {
		if group != "" {
			queryParams.Add("rule_group[]", group)
		}
	}

	// Add file paths filter if provided
	for _, file := range files {
		if file != "" {
			queryParams.Add("file[]", file)
		}
	}

	// Add exclude_alerts flag if true
	if excludeAlerts {
		queryParams.Add("exclude_alerts", "true")
	}

	// Add match label selectors if provided
	for _, match := range matchLabels {
		if match != "" {
			queryParams.Add("match[]", match)
		}
	}

	// Add group limit if provided
	if groupLimit != "" {
		queryParams.Add("group_limit", groupLimit)
	}

	// Combine base URL with query parameters if any
	requestURL := baseURL
	if len(queryParams) > 0 {
		requestURL = fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}

	// Send the request
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus rules: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check for error status
	if status, ok := result["status"].(string); ok && status != "success" {
		if errorType, ok := result["errorType"].(string); ok {
			if errorMsg, ok := result["error"].(string); ok {
				return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
			}
			return "", fmt.Errorf("Prometheus error: %s", errorType)
		}
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Format the result with additional information for no rules case
	if data, ok := result["data"].(map[string]interface{}); ok {
		if groups, ok := data["groups"].([]interface{}); ok && len(groups) == 0 {
			result["info"] = "No rules found matching the criteria."
		}
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}
