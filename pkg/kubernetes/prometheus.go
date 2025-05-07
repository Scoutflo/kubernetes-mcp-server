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
			// Make it clear this is a conclusive result, not a failure
			emptyResult := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "vector",
					"result":     []interface{}{},
				},
				"ERROR_TYPE": "NO_DATA_POINTS",
				"info":       "CONCLUSIVE RESULT: No data found for the query. This is a definitive answer, not an error - the query '" + query + "' executed successfully but returned no data points. This could mean the metric doesn't exist or no data points are available in the specified time range.",
				"query_used": query,
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
			// Make it clear this is a conclusive result, not a failure
			timeRange := fmt.Sprintf("from %s to %s with step %s",
				time.Unix(startTS, 0).Format(time.RFC3339),
				time.Unix(endTS, 0).Format(time.RFC3339),
				step)

			emptyResult := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "matrix",
					"result":     []interface{}{},
				},
				"ERROR_TYPE": "NO_DATA_POINTS_IN_RANGE",
				"info":       "CONCLUSIVE RESULT: No data found for the query. This is a definitive answer, not an error - the query '" + query + "' executed successfully but returned no data points in the specified time range " + timeRange + ". This could mean the metric doesn't exist or no data points are available in this range.",
				"query_used": query,
				"time_range": timeRange,
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

	// Check if metadata is empty (indicating the metric doesn't exist)
	if data, ok := metadataResult["data"].(map[string]interface{}); ok && len(data) == 0 {
		// Make it clear this is a conclusive result about the metric not existing
		emptyResult := map[string]interface{}{
			"name":       metricName,
			"status":     "success",
			"ERROR_TYPE": "METRIC_NOT_FOUND",
			"info":       "CONCLUSIVE RESULT: No metadata found for metric '" + metricName + "'. This is a definitive answer that this metric does not exist in Prometheus.",
			"metadata": map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{},
			},
		}

		resultJSON, err := json.MarshalIndent(emptyResult, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal empty result: %v", err)
		}
		return string(resultJSON), nil
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
		// Return a specific message for no data with ERROR_TYPE field to make it clearer
		emptyResult := map[string]interface{}{
			"status":        "success",
			"data":          []interface{}{},
			"ERROR_TYPE":    "NO_MATCHING_SERIES",
			"info":          "CONCLUSIVE RESULT: No series found matching the criteria. This means there are no metrics matching the provided selector patterns. This is a definitive answer, not a failure - the query executed successfully but found no matching series.",
			"selector_used": match,
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

// handlePrometheusResponse is a helper function to handle Prometheus API responses
func handlePrometheusResponse(bodyBytes []byte) (map[string]interface{}, error) {
	// Check for specific plain text error responses
	bodyStr := string(bodyBytes)
	if strings.Contains(bodyStr, "Method Not Allowed") {
		return nil, fmt.Errorf("Prometheus error: Method Not Allowed. The server does not support this operation or it's configured to reject it.")
	}

	// Try to parse the response as JSON
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		// If we can't parse as JSON, return the raw response for debugging
		return nil, fmt.Errorf("failed to decode response as JSON: %v, raw response: %s", err, bodyStr)
	}

	// Check for error status in the JSON response
	if status, ok := result["status"].(string); ok && status == "error" {
		errorType := "unknown"
		if et, ok := result["errorType"].(string); ok {
			errorType = et
		}

		errorMsg := "no error message provided"
		if em, ok := result["error"].(string); ok {
			errorMsg = em

			// Special handling for admin APIs disabled error
			if em == "admin APIs disabled" {
				// Just return the standard error message, the MCP handlers will add the JSON
				return nil, fmt.Errorf("Prometheus error: admin APIs are disabled on this server. This is a security configuration that prevents administrative operations.")
			}
		}

		return nil, fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
	}

	// For non-success status
	if status, ok := result["status"].(string); ok && status != "success" {
		return nil, fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	return result, nil
}

// CleanPrometheusTombstones removes tombstone files created during Prometheus data deletion operations
func (k *Kubernetes) CleanPrometheusTombstones() (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for admin endpoint to clean tombstones
	adminURL := fmt.Sprintf("%s/api/v1/admin/tsdb/clean_tombstones", prometheusURL)

	// Create a POST request
	req, err := http.NewRequest("POST", adminURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request to clean tombstones: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to clean Prometheus tombstones: %v", err)
	}
	defer resp.Body.Close()

	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Handle the response using the helper function
	result, err := handlePrometheusResponse(bodyBytes)
	if err != nil {
		return "", err
	}

	// Convert result to JSON string with success message
	result["message"] = "Prometheus tombstones have been successfully cleaned."
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// CreatePrometheusSnapshot creates a snapshot of the current Prometheus TSDB data
func (k *Kubernetes) CreatePrometheusSnapshot(skipHead bool) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for admin endpoint to create snapshot
	adminURL := fmt.Sprintf("%s/api/v1/admin/tsdb/snapshot", prometheusURL)

	// Add skip_head parameter if true
	if skipHead {
		adminURL = fmt.Sprintf("%s?skip_head=true", adminURL)
	}

	// Create a POST request
	req, err := http.NewRequest("POST", adminURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for snapshot: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create Prometheus snapshot: %v", err)
	}
	defer resp.Body.Close()

	// Read the entire response body first
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if this is an admin APIs disabled error
	bodyStr := string(bodyBytes)
	if strings.Contains(bodyStr, "admin APIs disabled") {
		// Return the exact error with the raw JSON response embedded
		return "", fmt.Errorf("admin API error: %s", bodyStr)
	}

	// Handle the response using the helper function
	result, err := handlePrometheusResponse(bodyBytes)
	if err != nil {
		return "", err
	}

	// Success case - Convert result to JSON string with additional information
	if data, ok := result["data"].(map[string]interface{}); ok {
		if name, ok := data["name"].(string); ok {
			result["message"] = fmt.Sprintf("Prometheus snapshot '%s' created successfully.", name)
		}
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// DeletePrometheusSeries deletes time series data matching specific criteria
func (k *Kubernetes) DeletePrometheusSeries(match []string, start, end *time.Time) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for admin endpoint to delete series
	baseURL := fmt.Sprintf("%s/api/v1/admin/tsdb/delete_series", prometheusURL)

	// Build query parameters
	queryParams := url.Values{}

	// Add match parameters (required)
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

	// Combine base URL with query parameters
	requestURL := fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())

	// Create a POST request
	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request to delete series: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to delete Prometheus series: %v", err)
	}
	defer resp.Body.Close()

	// Read the entire response body first
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for specific plain text error responses
	bodyStr := string(bodyBytes)
	if strings.Contains(bodyStr, "Method Not Allowed") {
		return "", fmt.Errorf("Prometheus error: Method Not Allowed. The server does not support this operation or it's configured to reject it.")
	}

	// Try to parse the response as JSON
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		// If we can't parse as JSON, return the raw response for debugging
		return "", fmt.Errorf("failed to decode response as JSON: %v, raw response: %s", err, bodyStr)
	}

	// Check for error status in the JSON response
	if status, ok := result["status"].(string); ok && status == "error" {
		errorType := "unknown"
		if et, ok := result["errorType"].(string); ok {
			errorType = et
		}

		errorMsg := "no error message provided"
		if em, ok := result["error"].(string); ok {
			errorMsg = em
		}

		return "", fmt.Errorf("Prometheus error (%s): %s", errorType, errorMsg)
	}

	// For non-success status
	if status, ok := result["status"].(string); ok && status != "success" {
		return "", fmt.Errorf("Prometheus returned non-success status: %s", status)
	}

	// Add success message to the result
	matchStr := strings.Join(match, ", ")
	timeInfo := ""
	if start != nil && !start.IsZero() {
		if end != nil && !end.IsZero() {
			timeInfo = fmt.Sprintf(" between %s and %s", start.String(), end.String())
		} else {
			timeInfo = fmt.Sprintf(" from %s onwards", start.String())
		}
	} else if end != nil && !end.IsZero() {
		timeInfo = fmt.Sprintf(" up to %s", end.String())
	}

	result["message"] = fmt.Sprintf("Successfully deleted time series matching: %s%s", matchStr, timeInfo)

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusAlertManagers retrieves information about Alertmanager instances known to Prometheus
func (k *Kubernetes) GetPrometheusAlertManagers() (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for alertmanagers endpoint
	alertManagersURL := fmt.Sprintf("%s/api/v1/alertmanagers", prometheusURL)

	// Send the request
	resp, err := http.Get(alertManagersURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus alertmanagers: %v", err)
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

	// Add a note if no alertmanagers are found
	if data, ok := result["data"].(map[string]interface{}); ok {
		activeCount := 0
		droppedCount := 0

		if active, ok := data["activeAlertmanagers"].([]interface{}); ok {
			activeCount = len(active)
		}

		if dropped, ok := data["droppedAlertmanagers"].([]interface{}); ok {
			droppedCount = len(dropped)
		}

		if activeCount == 0 && droppedCount == 0 {
			result["info"] = "No Alertmanager instances are currently connected to Prometheus."
		} else {
			result["info"] = fmt.Sprintf("Found %d active and %d dropped Alertmanager instances.",
				activeCount, droppedCount)
		}
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusRuntimeInfo retrieves detailed information about the Prometheus server's runtime state
func (k *Kubernetes) GetPrometheusRuntimeInfo() (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for runtime info endpoint
	runtimeInfoURL := fmt.Sprintf("%s/api/v1/status/runtimeinfo", prometheusURL)

	// Send the request
	resp, err := http.Get(runtimeInfoURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus runtime info: %v", err)
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

// GetPrometheusTSDBStatus retrieves information about the time series database (TSDB) status in Prometheus
func (k *Kubernetes) GetPrometheusTSDBStatus(limit int) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for TSDB status endpoint
	tsdbStatusURL := fmt.Sprintf("%s/api/v1/status/tsdb", prometheusURL)

	// Add limit parameter if provided
	if limit > 0 {
		tsdbStatusURL = fmt.Sprintf("%s?limit=%d", tsdbStatusURL, limit)
	}

	// Send the request
	resp, err := http.Get(tsdbStatusURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus TSDB status: %v", err)
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

// GetPrometheusWALReplayStatus retrieves the status of Write-Ahead Log (WAL) replay operations in Prometheus
func (k *Kubernetes) GetPrometheusWALReplayStatus() (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for WAL replay status endpoint
	walReplayURL := fmt.Sprintf("%s/api/v1/status/walreplay", prometheusURL)

	// Send the request
	resp, err := http.Get(walReplayURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus WAL replay status: %v", err)
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
