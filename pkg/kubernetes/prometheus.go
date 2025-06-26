package kubernetes

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/scoutflo/kubernetes-mcp-server/pkg/llm"
)

// QueryPrometheus sends an instant query to Prometheus
func (k *Kubernetes) QueryPrometheus(query string, queryTime *time.Time, timeout string) (string, error) {
	// Build the endpoint with query parameters
	endpoint := fmt.Sprintf("/apis/v1/prometheus/query-metrics?query=%s", url.QueryEscape(query))

	// Add timestamp if provided
	if queryTime != nil && !queryTime.IsZero() {
		endpoint = fmt.Sprintf("%s&time=%d", endpoint, queryTime.Unix())
	}

	// Add timeout if provided
	if timeout != "" {
		endpoint = fmt.Sprintf("%s&timeout=%s", endpoint, url.QueryEscape(timeout))
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus: %w", err)
	}

	return string(response), nil
}

// QueryPrometheusRange sends a range query to Prometheus
func (k *Kubernetes) QueryPrometheusRange(query string, start, end time.Time, step string, timeout string) (string, error) {
	// Convert times to Unix timestamps
	startTS := start.Unix()
	endTS := end.Unix()

	// Build the endpoint with query parameters
	endpoint := fmt.Sprintf("/apis/v1/prometheus/query-metrics-range?query=%s&start=%d&end=%d&step=%s",
		url.QueryEscape(query), startTS, endTS, url.QueryEscape(step))

	// Add timeout if provided
	if timeout != "" {
		endpoint = fmt.Sprintf("%s&timeout=%s", endpoint, url.QueryEscape(timeout))
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus range: %w", err)
	}

	return string(response), nil
}

// ListPrometheusMetrics retrieves all available metrics from Prometheus
func (k *Kubernetes) ListPrometheusMetrics() (string, error) {
	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", "/apis/v1/prometheus/list-metrics", nil)
	if err != nil {
		return "", fmt.Errorf("failed to list Prometheus metrics: %w", err)
	}

	return string(response), nil
}

// GetPrometheusMetricInfo retrieves detailed information about a specific metric
func (k *Kubernetes) GetPrometheusMetricInfo(metricName string, includeStats bool) (string, error) {
	// Build the endpoint with metric parameter
	endpoint := fmt.Sprintf("/apis/v1/prometheus/metric-info?metric=%s", url.QueryEscape(metricName))

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus metric info: %w", err)
	}

	return string(response), nil
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
	// Build request body for POST request
	reqBody := map[string]interface{}{
		"match": match,
	}

	// Add start time if provided
	if start != nil && !start.IsZero() {
		reqBody["start"] = fmt.Sprintf("%d", start.Unix())
	}

	// Add end time if provided
	if end != nil && !end.IsZero() {
		reqBody["end"] = fmt.Sprintf("%d", end.Unix())
	}

	// Add limit if provided and valid
	if limit > 0 {
		reqBody["limit"] = limit
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "/apis/v1/prometheus/series", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus series: %w", err)
	}

	return string(response), nil
}

// GetPrometheusTargets retrieves information about Prometheus targets
func (k *Kubernetes) GetPrometheusTargets(state, scrapePool string) (string, error) {
	// Build endpoint with optional query parameters
	endpoint := "/apis/v1/prometheus/targets"
	var params []string

	// Add state filter if provided
	if state != "" {
		params = append(params, fmt.Sprintf("state=%s", url.QueryEscape(state)))
	}

	// Add scrape pool filter if provided
	if scrapePool != "" {
		params = append(params, fmt.Sprintf("scrape_pool=%s", url.QueryEscape(scrapePool)))
	}

	// Add query parameters if any
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus targets: %w", err)
	}

	return string(response), nil
}

// GetPrometheusTargetMetadata retrieves metadata about metrics exposed by targets
func (k *Kubernetes) GetPrometheusTargetMetadata(matchTarget, metric string, limit int) (string, error) {
	// Build endpoint with optional query parameters
	endpoint := "/apis/v1/prometheus/targets-metadata"
	var params []string

	// Add match_target parameter if provided
	if matchTarget != "" {
		params = append(params, fmt.Sprintf("match_target=%s", url.QueryEscape(matchTarget)))
	}

	// Add metric parameter if provided
	if metric != "" {
		params = append(params, fmt.Sprintf("metric=%s", url.QueryEscape(metric)))
	}

	// Add limit parameter if provided and valid
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}

	// Add query parameters if any
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus target metadata: %w", err)
	}

	return string(response), nil
}

// GetPrometheusAlerts retrieves all currently firing alerts from Prometheus
func (k *Kubernetes) GetPrometheusAlerts() (string, error) {
	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", "/apis/v1/prometheus/alerts", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus alerts: %w", err)
	}

	return string(response), nil
}

// GetPrometheusRules retrieves information about configured alerting and recording rules
func (k *Kubernetes) GetPrometheusRules(groupLimit string, ruleNames, ruleGroups, files []string, excludeAlerts bool, matchLabels []string) (string, error) {
	// Build request body for POST request
	reqBody := map[string]interface{}{}

	// Add rule names filter if provided
	if len(ruleNames) > 0 {
		reqBody["rule_name"] = ruleNames
	}

	// Add rule group filter if provided
	if len(ruleGroups) > 0 {
		reqBody["rule_group"] = ruleGroups
	}

	// Add file paths filter if provided
	if len(files) > 0 {
		reqBody["file"] = files
	}

	// Add exclude_alerts if provided
	if excludeAlerts {
		reqBody["exclude_alerts"] = []string{} // Empty array to enable exclude_alerts
	}

	// Add group limit if provided
	if groupLimit != "" {
		reqBody["group_limit"] = groupLimit
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "/apis/v1/prometheus/rules", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus rules: %w", err)
	}

	return string(response), nil
}

// GetPrometheusRuntimeInfo retrieves detailed information about the Prometheus server's runtime state
func (k *Kubernetes) GetPrometheusRuntimeInfo() (string, error) {
	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", "/apis/v1/prometheus/runtime-info", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus runtime info: %w", err)
	}

	return string(response), nil
}

// GetPrometheusTSDBStatus retrieves information about the time series database (TSDB) status in Prometheus
func (k *Kubernetes) GetPrometheusTSDBStatus(limit int) (string, error) {
	// Build endpoint with optional limit parameter
	endpoint := "/apis/v1/prometheus/TSDB-status"

	// Add limit parameter if provided and valid
	if limit > 0 {
		endpoint = fmt.Sprintf("%s?limit=%d", endpoint, limit)
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus TSDB status: %w", err)
	}

	return string(response), nil
}

// ListPrometheusLabelNames retrieves all available label names from Prometheus
func (k *Kubernetes) ListPrometheusLabelNames(startRfc3339, endRfc3339 string, limit int, matches []string) (string, error) {
	// Build request body for POST request
	reqBody := map[string]interface{}{}

	// Add start time if provided
	if startRfc3339 != "" {
		reqBody["startRfc3339"] = startRfc3339
	}

	// Add end time if provided
	if endRfc3339 != "" {
		reqBody["endRfc3339"] = endRfc3339
	}

	// Add limit if provided and valid
	if limit > 0 {
		reqBody["limit"] = limit
	}

	// Add match parameters if provided
	if len(matches) > 0 {
		reqBody["matches"] = matches
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "/apis/v1/prometheus/list-label-names", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus label names: %w", err)
	}

	return string(response), nil
}

// ListPrometheusLabelValues retrieves all values for a specific label name from Prometheus
func (k *Kubernetes) ListPrometheusLabelValues(labelName, startRfc3339, endRfc3339 string, limit int, matches []string) (string, error) {
	// Build request body for POST request
	reqBody := map[string]interface{}{}

	// Add start time if provided
	if startRfc3339 != "" {
		reqBody["startRfc3339"] = startRfc3339
	}

	// Add end time if provided
	if endRfc3339 != "" {
		reqBody["endRfc3339"] = endRfc3339
	}

	// Add limit if provided and valid
	if limit > 0 {
		reqBody["limit"] = limit
	}

	// Add match parameters if provided
	if len(matches) > 0 {
		reqBody["matches"] = matches
	}

	// Make API request to K8s Dashboard using the label name in the URL
	endpoint := fmt.Sprintf("/apis/v1/prometheus/labels/%s/values", url.PathEscape(labelName))
	response, err := k.MakeAPIRequest("POST", endpoint, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus label values: %w", err)
	}

	return string(response), nil
}

// GetPrometheusStatus retrieves Prometheus server status configuration
func (k *Kubernetes) GetPrometheusStatus() (string, error) {
	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("GET", "/apis/v1/prometheus/status", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus status: %w", err)
	}

	return string(response), nil
}

// CreatePrometheusAlert creates a new Prometheus alert rule via K8s Dashboard API
func (k *Kubernetes) CreatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration string, annotations, alertLabels map[string]string) (string, error) {
	// Build request body
	reqBody := map[string]interface{}{
		"alert_name": alertName,
		"expression": expression,
		"app_label":  appLabel,
		"namespace":  namespace,
	}

	// Add optional parameters
	if interval != "" {
		reqBody["interval"] = interval
	}
	if forDuration != "" {
		reqBody["for"] = forDuration
	}
	if annotations != nil && len(annotations) > 0 {
		reqBody["annotations"] = annotations
	}
	if alertLabels != nil && len(alertLabels) > 0 {
		reqBody["alert_labels"] = alertLabels
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "/apis/v1/prometheus/create-alert", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create Prometheus alert: %w", err)
	}

	return string(response), nil
}

// UpdatePrometheusAlert updates an existing Prometheus alert rule via K8s Dashboard API
func (k *Kubernetes) UpdatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration string, annotations, alertLabels map[string]string) (string, error) {
	// Build request body
	reqBody := map[string]interface{}{
		"alert_name": alertName,
		"app_label":  appLabel,
		"namespace":  namespace,
	}

	// Add optional parameters only if they are not empty
	if expression != "" {
		reqBody["expression"] = expression
	}
	if interval != "" {
		reqBody["interval"] = interval
	}
	if forDuration != "" {
		reqBody["for"] = forDuration
	}
	if annotations != nil && len(annotations) > 0 {
		reqBody["annotations"] = annotations
	}
	if alertLabels != nil && len(alertLabels) > 0 {
		reqBody["alert_labels"] = alertLabels
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "/apis/v1/prometheus/update-alert", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to update Prometheus alert: %w", err)
	}

	return string(response), nil
}

// DeletePrometheusAlert deletes a Prometheus alert rule via K8s Dashboard API
func (k *Kubernetes) DeletePrometheusAlert(appLabel, namespace, alertName string) (string, error) {
	// Build request body
	reqBody := map[string]interface{}{
		"app_label": appLabel,
		"namespace": namespace,
	}

	// Add alert name if specified (optional - if not provided, deletes entire rule group)
	if alertName != "" {
		reqBody["alert_name"] = alertName
	}

	// Make API request to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "/apis/v1/prometheus/delete-alert", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to delete Prometheus alert: %w", err)
	}

	return string(response), nil
}

// NOTE: The following functions are removed as they are admin operations
// that should not be exposed through the MCP server for security reasons:
// - CleanPrometheusTombstones
// - CreatePrometheusSnapshot
// - DeletePrometheusSeries
// - GetPrometheusAlertManagers
// - GetPrometheusWALReplayStatus
