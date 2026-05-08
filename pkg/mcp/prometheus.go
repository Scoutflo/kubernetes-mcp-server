package mcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initPrometheus() []server.ServerTool {
	return []server.ServerTool{
		{Tool: WithMeta(
			mcp.NewTool("prometheus_generate_query",
				mcp.WithDescription("Convert natural language descriptions into valid PromQL queries. Use when the correct metric name or query structure is unknown — generates a query to validate with prometheus_metrics_query or prometheus_metrics_query_range."),
				mcp.WithString("description", mcp.Description("Natural language description of the metric you want to query"), mcp.Required()),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusGenerateQuery},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_metrics_query",
				mcp.WithDescription("Execute an instant PromQL query returning current metric values. Validate the metric name with prometheus_metric_info before querying if uncertain. Use prometheus_metrics_query_range for trend analysis over time."),
				mcp.WithString("query", mcp.Description("Prometheus PromQL expression query string"), mcp.Required()),
				mcp.WithString("time", mcp.Description("Evaluation timestamp in RFC3339 or unix timestamp format (optional)")),
				mcp.WithString("timeout", mcp.Description("Evaluation timeout (optional)")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusMetrics},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_metrics_query_range",
				mcp.WithDescription("Execute a PromQL range query returning time-series data. Always narrow to the incident window first — use range (e.g., '1h') or explicit start/end to avoid oversized responses. Step is auto-selected when using range; specify explicitly when start/end are provided. Validate metric names with prometheus_metric_info before querying if uncertain."),
				mcp.WithString("query", mcp.Description("Prometheus PromQL expression query string"), mcp.Required()),
				mcp.WithString("start", mcp.Description("Start timestamp in RFC3339 or Unix timestamp format")),
				mcp.WithString("end", mcp.Description("End timestamp in RFC3339 or Unix timestamp format")),
				mcp.WithString("step", mcp.Description("Query resolution step width (e.g., '15s', '1m', '1h')")),
				mcp.WithString("range", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') - alternative to start/end")),
				mcp.WithString("timeout", mcp.Description("Evaluation timeout (optional)")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusMetricsRange},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_list_metrics",
				mcp.WithDescription("List all available metric names in Prometheus. Use to discover exact metric names before querying — metric names must be exact; guessed names will return no data. Call before prometheus_metric_info when the metric name is completely unknown."),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusListMetrics},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_metric_info",
				mcp.WithDescription("Retrieve metadata (type, help text, unit) and optional statistics for a specific metric. Use to validate a metric name before querying and to understand its label set. Prefer over prometheus_list_metrics when the approximate metric name is already known."),
				mcp.WithString("metric", mcp.Description("Name of the metric to get information about"), mcp.Required()),
				mcp.WithBoolean("include_statistics", mcp.Description("Include count, min, max, and avg statistics for this metric. May be slower for metrics with many time series.")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusMetricInfo},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_series_query",
				mcp.WithDescription("Find time series matching label selectors. Use to discover which label combinations exist for a metric before constructing targeted PromQL queries. Call prometheus_list_label_names first when label key names are unknown."),
				mcp.WithArray("match", mcp.Description("Series selector arguments"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
					mcp.Required()),
				mcp.WithString("start", mcp.Description("Start timestamp in RFC3339 or Unix timestamp format (optional)")),
				mcp.WithString("end", mcp.Description("End timestamp in RFC3339 or Unix timestamp format (optional)")),
				mcp.WithString("time_window", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') — alternative to start/end")),
				mcp.WithNumber("limit", mcp.Description("Maximum number of returned items (optional)")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusSeries},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_targets",
				mcp.WithDescription("List active Prometheus scrape targets with health status and last scrape error. Use to diagnose missing metrics — if a target is down or has scrape errors, queries for its metrics will return no data. Filter by state='active' to focus on healthy targets."),
				mcp.WithString("state", mcp.Description("Target state filter, must be one of: active, dropped, any (optional)")),
				mcp.WithString("scrape_pool", mcp.Description("Scrape pool name (optional)")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusTargets},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_targets_metadata",
				mcp.WithDescription("Retrieve metric metadata from Prometheus scrape targets. Returns metadata about metrics exposed by specific targets including help text, type, and unit information. Use when you need to verify what metrics a target exposes or validate metric consistency across targets. Targets are Prometheus scrape endpoints that expose metrics."),
				mcp.WithString("match_target", mcp.Description("Target label selectors (optional)")),
				mcp.WithString("metric", mcp.Description("Metric name (optional)")),
				mcp.WithNumber("limit", mcp.Description("Maximum number of targets (optional)")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusTargetMetadata},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_list_label_names",
				mcp.WithDescription("List all label names available in Prometheus. Call before prometheus_list_label_values to discover valid label key names — do not guess label keys. Use matches filter to scope to a specific metric."),
				mcp.WithString("start", mcp.Description("Optionally, the start time of the time range to filter the results by in RFC3339 or Unix timestamp format")),
				mcp.WithString("end", mcp.Description("Optionally, the end time of the time range to filter the results by in RFC3339 or Unix timestamp format")),
				mcp.WithNumber("limit", mcp.Description("Optionally, the maximum number of results to return")),
				mcp.WithArray("matches", mcp.Description("Optionally, a list of label matchers to filter the results by"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusListLabelNames},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_list_label_values",
				mcp.WithDescription("List all values for a specific label name. Call after prometheus_list_label_names to confirm the label key exists. Essential for resolving pod, namespace, service, or instance identifiers before constructing incident-scoped PromQL queries."),
				mcp.WithString("labelName", mcp.Description("The name of the label to query"), mcp.Required()),
				mcp.WithString("start", mcp.Description("Optionally, the start time of the query in RFC3339 or Unix timestamp format")),
				mcp.WithString("end", mcp.Description("Optionally, the end time of the query in RFC3339 or Unix timestamp format")),
				mcp.WithNumber("limit", mcp.Description("Optionally, the maximum number of results to return")),
				mcp.WithArray("matches", mcp.Description("Optionally, a list of selectors to filter the results by"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusListLabelValues},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_get_alerts",
				mcp.WithDescription("List currently firing Prometheus alerts. First call for any incident — reveals active alert names, labels (pod/namespace/service), severity, and active duration. Use these labels to scope subsequent metric queries. Scope with time_window or start_time/end_time to match an incident window; defaults to the last 1h."),
				mcp.WithString("start_time", mcp.Description("Start time in RFC3339 format or Unix timestamp. Required if end_time is provided.")),
				mcp.WithString("end_time", mcp.Description("End time in RFC3339 format or Unix timestamp. Required if start_time is provided.")),
				mcp.WithString("time_window", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') — alternative to start_time/end_time.")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusGetAlerts},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_get_rules",
				mcp.WithDescription("Retrieve configured alerting and recording rules. Use to inspect the PromQL expression, 'for' duration, and threshold behind a firing alert — essential for understanding why an alert triggered. Filter by rule_name when the alert name is known. Scope with time_window or start_time/end_time to match an incident window; defaults to the last 1h."),
				mcp.WithArray("rule_name", mcp.Description("Rule names filter"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
				mcp.WithArray("rule_group", mcp.Description("Rule group names filter"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
				mcp.WithArray("file", mcp.Description("File paths filter"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
				mcp.WithBoolean("exclude_alerts", mcp.Description("If true, exclude alerting rules from results and return only recording rules")),
				mcp.WithArray("match", mcp.Description("Label matchers to filter rules, e.g. [\"severity=critical\", \"team=backend\"]"),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
				mcp.WithNumber("group_limit", mcp.Description("Maximum number of rule groups to return (0 for unlimited)")),
				mcp.WithString("start_time", mcp.Description("Start time in RFC3339 format or Unix timestamp. Required if end_time is provided.")),
				mcp.WithString("end_time", mcp.Description("End time in RFC3339 format or Unix timestamp. Required if start_time is provided.")),
				mcp.WithString("time_window", mcp.Description("Time range from now (e.g., '1h', '24h', '7d') — alternative to start_time/end_time.")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusGetRules},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_create_alert",
				mcp.WithDescription("Create a new Prometheus alert rule as a PrometheusRule CRD in Kubernetes. Write operation. Always validate the PromQL expression with prometheus_metrics_query before creating — use prometheus_generate_query if the expression is uncertain. Requires namespace and applabel."),
				mcp.WithString("alertname", mcp.Description("Name of the alert to create"), mcp.Required()),
				mcp.WithString("expression", mcp.Description("PromQL expression that defines the alert condition, If not provided, please generate a query using prometheus_generate_query tool"), mcp.Required()),
				mcp.WithString("applabel", mcp.Description("Application label used to identify the PrometheusRule resource, use alertname if applabel is not provided"), mcp.Required()),
				mcp.WithString("namespace", mcp.Description("Kubernetes namespace to create the alert in"), mcp.Required()),
				mcp.WithString("interval", mcp.Description("Evaluation interval for the alert group (e.g., '30s', '1m', '5m')")),
				mcp.WithString("for", mcp.Description("Duration for which the condition must be true before firing (e.g., '5m')")),
				mcp.WithObject("annotations", mcp.Description("JSON object of alert annotations as string key-value pairs, e.g. {\"summary\": \"High CPU usage\", \"description\": \"CPU usage is above 80% for pod {{ $labels.pod }}\"}")),
				mcp.WithObject("alertlabels", mcp.Description("JSON object of labels to attach to the fired alert as string key-value pairs, e.g. {\"severity\": \"critical\", \"team\": \"backend\"}")),
			),
			map[string]any{
				"provider": ProviderPrometheus,
				"hitl": map[string]any{
					"required":     true,
					"riskLevel":    RiskLow,
					"approvalType": "single",
					"message":      "This will create a new Prometheus alert rule. Proceed?",
				},
			},
		), Handler: s.prometheusCreateAlert},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_update_alert",
				mcp.WithDescription("Update an existing Prometheus alert rule. Write operation. Validate the new PromQL expression with prometheus_metrics_query before updating. Only provide fields you want to change — omitted fields retain their current values."),
				mcp.WithString("alertname", mcp.Description("Name of the alert to update"), mcp.Required()),
				mcp.WithString("applabel", mcp.Description("Application label that identifies the PrometheusRule resource, use alertname if applabel is not provided"), mcp.Required()),
				mcp.WithString("namespace", mcp.Description("Kubernetes namespace of the alert"), mcp.Required()),
				mcp.WithString("expression", mcp.Description("New PromQL expression for the alert condition")),
				mcp.WithString("interval", mcp.Description("New evaluation interval for the alert group (e.g., '30s', '1m', '5m')")),
				mcp.WithString("for", mcp.Description("New duration for which the condition must be true before firing (e.g., '5m')")),
				mcp.WithObject("annotations", mcp.Description("JSON object of new or updated alert annotations as string key-value pairs, e.g. {\"summary\": \"High CPU usage\", \"description\": \"CPU above threshold\"}")),
				mcp.WithObject("alertlabels", mcp.Description("JSON object of new or updated labels for the alert as string key-value pairs, e.g. {\"severity\": \"warning\", \"team\": \"platform\"}")),
			),
			map[string]any{
				"provider": ProviderPrometheus,
				"hitl": map[string]any{
					"required":     true,
					"riskLevel":    RiskLow,
					"approvalType": "single",
					"message":      "This will update an existing Prometheus alert rule. Proceed?",
				},
			},
		), Handler: s.prometheusUpdateAlert},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_delete_alert",
				mcp.WithDescription("Delete a Prometheus alert rule. Write operation — irreversible. Deletes the entire PrometheusRule resource identified by applabel and namespace unless alertname is specified (which removes only that alert from the group)."),
				mcp.WithString("applabel", mcp.Description("Application label that identifies the PrometheusRule resource, use alertname if applabel is not provided"), mcp.Required()),
				mcp.WithString("namespace", mcp.Description("Kubernetes namespace of the alert"), mcp.Required()),
				mcp.WithString("alertname", mcp.Description("Name of the specific alert to delete within the rule group (optional)")),
			),
			map[string]any{
				"provider": ProviderPrometheus,
				"hitl": map[string]any{
					"required":     true,
					"riskLevel":    RiskLow,
					"approvalType": "single",
					"message":      "This will delete a Prometheus alert rule. Proceed?",
				},
			},
		), Handler: s.prometheusDeleteAlert},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_runtimeinfo",
				mcp.WithDescription("Retrieve Prometheus server runtime information including version, storage path, retention settings, and uptime. Use to diagnose Prometheus health issues or verify configuration state when targets or queries behave unexpectedly."),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusRuntimeInfo},
		{Tool: WithMeta(
			mcp.NewTool("prometheus_TSDB_status",
				mcp.WithDescription("Retrieve TSDB (Time Series Database) status information from Prometheus. Returns storage statistics including head stats, series data, chunk information, and retention details. Use when you need to monitor Prometheus storage health, check data retention, or diagnose storage issues. TSDB is Prometheus's storage backend."),
				mcp.WithNumber("limit", mcp.Description("Number of items limit")),
			),
			map[string]any{"provider": ProviderPrometheus},
		), Handler: s.prometheusTSDBStatus},
	}
}

// prometheusMetrics handles the prometheus_metrics_query tool request
func (s *Server) prometheusMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()

	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_metrics failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	queryArg := ctr.GetString("query", "")
	timeArg := ctr.GetString("time", "")
	timeWindowStr := ctr.GetString("time_window", "")
	timeout := ctr.GetString("timeout", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool call: prometheus_metrics_query - query=%s, time=%s, time_window=%s, timeout=%s - got called by session id: %s", queryArg, timeArg, timeWindowStr, timeout, sessionID)

	if queryArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query failed after %v: missing required parameter: query by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	query := queryArg

	// Validate timeout format if provided
	if timeout != "" {
		if _, err := time.ParseDuration(timeout); err != nil {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_metrics_query failed after %v: invalid timeout format: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("invalid timeout format '%s': must be a valid duration string (e.g., '30s', '1m', '5m'): %v", timeout, err)), nil
		}
	}

	// Extract time parameter - either time or time_window must be provided, default to 1h if neither provided
	var queryTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_metrics_query failed after %v: invalid time_window format: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		evalTime := now.Add(-duration)
		queryTime = &evalTime
	} else if timeArg != "" {
		parsedTime := parseTime(timeArg, time.Time{})
		if parsedTime.IsZero() {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_metrics_query failed after %v: invalid time format by session id: %s", duration, sessionID)
			return NewTextResult("", errors.New("invalid time format, use RFC3339 (e.g., '2024-01-01T00:00:00Z') or Unix timestamp")), nil
		}
		queryTime = &parsedTime
	} else {
		// Default to 1 hour ago if neither time nor time_window is provided
		now := time.Now()
		defaultTime := now.Add(-1 * time.Hour)
		queryTime = &defaultTime
		klog.V(1).Infof("Tool call: prometheus_metrics_query - using default time_window of 1h by session id: %s", sessionID)
	}

	// Execute the instant query with the provided parameters
	ret, err := k.QueryPrometheus(query, queryTime, timeout)
	if err != nil {
		duration := time.Since(start)
		errMsg := err.Error()
		// Check for common error patterns and provide more helpful messages
		if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
			klog.Errorf("Tool call: prometheus_metrics_query failed after %v: metric not found: %s by session id: %s", duration, query, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Metric not found. The specified metric '%s' does not exist in Prometheus. Please check the metric name and ensure it's correctly spelled.", query)), nil
		} else if strings.Contains(errMsg, "parse error") {
			klog.Errorf("Tool call: prometheus_metrics_query failed after %v: invalid PromQL syntax: %s by session id: %s", duration, query, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Invalid PromQL query syntax in '%s'. Please check your query format.", query)), nil
		} else if strings.Contains(errMsg, "failed to discover Prometheus") {
			klog.Errorf("Tool call: prometheus_metrics_query failed after %v: cannot connect to Prometheus server by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
		}
		klog.Errorf("Tool call: prometheus_metrics_query failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: Failed to execute Prometheus query: %v", err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_metrics_query completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusMetricsRange handles the prometheus_metrics_query_range tool request
func (s *Server) prometheusMetricsRange(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_metrics_query_range failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	queryArg := ctr.GetString("query", "")
	startArg := ctr.GetString("start", "")
	endArg := ctr.GetString("end", "")
	stepArg := ctr.GetString("step", "")
	rangeArg := ctr.GetString("range", "")
	timeout := ctr.GetString("timeout", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool call: prometheus_metrics_query_range - query=%s, start=%s, end=%s, step=%s, range=%s, timeout=%s - got called by session id: %s",
		queryArg, startArg, endArg, stepArg, rangeArg, timeout, sessionID)

	// Validate required parameters
	if queryArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: query by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}

	// Validate timeout format if provided
	if timeout != "" {
		if _, err := time.ParseDuration(timeout); err != nil {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid timeout format: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("invalid timeout format '%s': must be a valid duration string (e.g., '30s', '1m', '5m'): %v", timeout, err)), nil
		}
	}

	// using range parameter as alternative to start/end/step
	if rangeArg != "" {
		// Parse range duration
		rangeDuration, err := time.ParseDuration(rangeArg)
		if err != nil {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid range format: %s by session id: %s", duration, rangeArg, sessionID)
			return NewTextResult("", fmt.Errorf("invalid range format '%s': %v", rangeArg, err)), nil
		}

		// Set default values based on range
		endTime := time.Now()
		startTime := endTime.Add(-rangeDuration)

		// Auto-determine step based on range duration to limit data points
		// Goal: Keep response size manageable by limiting to ~60-170 points per series
		step := "1m" // default
		if rangeDuration <= time.Hour {
			step = "1m" // 60 points max
		} else if rangeDuration <= 6*time.Hour {
			step = "5m" // 72 points max
		} else if rangeDuration <= 24*time.Hour {
			step = "15m" // 96 points max
		} else if rangeDuration <= 7*24*time.Hour {
			step = "1h" // 168 points max
		} else {
			step = "6h"
		}

		// Override with provided values if specified
		if startArg != "" {
			startTime = parseTime(startArg, startTime)
			if startTime.IsZero() {
				duration := time.Since(start)
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid start time format: %s by session id: %s", duration, startArg, sessionID)
				return NewTextResult("", errors.New("invalid start time format")), nil
			}
		}
		if endArg != "" {
			endTime = parseTime(endArg, endTime)
			if endTime.IsZero() {
				duration := time.Since(start)
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid end time format: %s by session id: %s", duration, endArg, sessionID)
				return NewTextResult("", errors.New("invalid end time format")), nil
			}
		}
		if stepArg != "" {
			// Validate step format
			if _, err := time.ParseDuration(stepArg); err != nil {
				duration := time.Since(start)
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid step format: %v by session id: %s", duration, err, sessionID)
				return NewTextResult("", fmt.Errorf("invalid step format '%s': must be a valid duration string (e.g., '15s', '1m', '1h'): %v", stepArg, err)), nil
			}
			step = stepArg
		}

		// Validate start < end
		if startTime.After(endTime) || startTime.Equal(endTime) {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: start time must be before end time by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("start time (%s) must be before end time (%s)", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))), nil
		}

		ret, err := k.QueryPrometheusRange(queryArg, startTime, endTime, step, timeout)
		if err != nil {
			duration := time.Since(start)
			errMsg := err.Error()
			if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: metric not found: %s by session id: %s", duration, queryArg, sessionID)
				return NewTextResult("", fmt.Errorf("ERROR: Metric not found. The specified metric '%s' does not exist in Prometheus. Please check the metric name and ensure it's correctly spelled.", queryArg)), nil
			} else if strings.Contains(errMsg, "parse error") {
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid PromQL syntax: %s by session id: %s", duration, queryArg, sessionID)
				return NewTextResult("", fmt.Errorf("ERROR: Invalid PromQL query syntax in '%s'. Please check your query format.", queryArg)), nil
			} else if strings.Contains(errMsg, "failed to discover Prometheus") {
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: cannot connect to Prometheus server by session id: %s", duration, sessionID)
				return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
			}
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Failed to execute Prometheus range query: %v", err)), nil
		}

		// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
		if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
			ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
		}

		duration := time.Since(start)
		klog.V(1).Infof("Tool call: prometheus_metrics_query_range completed successfully in %v by session id: %s", duration, sessionID)
		return NewTextResult(ret, nil), nil
	}

	// validation for start/end/step when range is not provided
	// If none are provided, default to 1h range (reduced from 24h to prevent large responses)
	if startArg == "" && endArg == "" && stepArg == "" {
		// Default to 1 hour ago to now
		endTime := time.Now()
		startTime := endTime.Add(-1 * time.Hour)
		step := "1m" // Default step for 1h range (60 data points max)
		klog.V(1).Infof("Tool call: prometheus_metrics_query_range - using default range of 1h by session id: %s", sessionID)
		ret, err := k.QueryPrometheusRange(queryArg, startTime, endTime, step, timeout)
		if err != nil {
			duration := time.Since(start)
			errMsg := err.Error()
			if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: metric not found: %s by session id: %s", duration, queryArg, sessionID)
				return NewTextResult("", fmt.Errorf("ERROR: Metric not found. The specified metric '%s' does not exist in Prometheus. Please check the metric name and ensure it's correctly spelled.", queryArg)), nil
			} else if strings.Contains(errMsg, "parse error") {
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid PromQL syntax: %s by session id: %s", duration, queryArg, sessionID)
				return NewTextResult("", fmt.Errorf("ERROR: Invalid PromQL query syntax in '%s'. Please check your query format.", queryArg)), nil
			} else if strings.Contains(errMsg, "failed to discover Prometheus") {
				klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: cannot connect to Prometheus server by session id: %s", duration, sessionID)
				return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
			}
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Failed to execute Prometheus range query: %v", err)), nil
		}
		if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
			ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
		}
		duration := time.Since(start)
		klog.V(1).Infof("Tool call: prometheus_metrics_query_range completed successfully in %v by session id: %s", duration, sessionID)
		return NewTextResult(ret, nil), nil
	}

	if startArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: start (or use range parameter) by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: start (or use range parameter)")), nil
	}
	if endArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: end (or use range parameter) by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: end (or use range parameter)")), nil
	}
	if stepArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: step (or use range parameter) by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: step (or use range parameter)")), nil
	}

	// Parse query
	query := queryArg

	// Parse start time
	startTime := parseTime(startArg, time.Now().Add(-1*time.Hour))
	if startTime.IsZero() {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid start time format: %s by session id: %s", duration, startArg, sessionID)
		return NewTextResult("", errors.New("invalid start time format")), nil
	}

	// Parse end time
	endTime := parseTime(endArg, time.Now())
	if endTime.IsZero() {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid end time format: %s by session id: %s", duration, endArg, sessionID)
		return NewTextResult("", errors.New("invalid end time format")), nil
	}

	// Parse and validate step
	step := stepArg
	if _, err := time.ParseDuration(step); err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid step format: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("invalid step format '%s': must be a valid duration string (e.g., '15s', '1m', '1h'): %v", step, err)), nil
	}

	// Validate start < end
	if startTime.After(endTime) || startTime.Equal(endTime) {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: start time must be before end time by session id: %s", duration, sessionID)
		return NewTextResult("", fmt.Errorf("start time (%s) must be before end time (%s)", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))), nil
	}

	// Execute the range query with the provided parameters
	ret, err := k.QueryPrometheusRange(query, startTime, endTime, step, timeout)
	if err != nil {
		duration := time.Since(start)
		errMsg := err.Error()
		// Check for common error patterns and provide more helpful messages
		if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: metric not found: %s by session id: %s", duration, query, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Metric not found. The specified metric '%s' does not exist in Prometheus. Please check the metric name and ensure it's correctly spelled.", query)), nil
		} else if strings.Contains(errMsg, "parse error") {
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid PromQL syntax: %s by session id: %s", duration, query, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Invalid PromQL query syntax in '%s'. Please check your query format.", query)), nil
		} else if strings.Contains(errMsg, "failed to discover Prometheus") {
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: cannot connect to Prometheus server by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
		} else if strings.Contains(errMsg, "invalid step") {
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: invalid step parameter: %s by session id: %s", duration, step, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Invalid step parameter '%s'. Step must be a valid duration (e.g., '15s', '1m', '1h').", step)), nil
		} else if strings.Contains(errMsg, "resolution") || strings.Contains(errMsg, "step") {
			klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: step parameter issue: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Step parameter issue: %v. Adjust the step size or time range.", err)), nil
		}
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: Failed to execute Prometheus range query: %v", err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_metrics_query_range completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusListMetrics handles the prometheus_list_metrics tool request
func (s *Server) prometheusListMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_list_metrics failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_list_metrics - got called by session id: %s", sessionID)

	ret, err := k.ListPrometheusMetrics()
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_list_metrics failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list Prometheus metrics: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_list_metrics completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusMetricInfo handles the prometheus_metric_info tool request
func (s *Server) prometheusMetricInfo(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_metric_info failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	metric := ctr.GetString("metric", "")
	includeStats := ctr.GetBool("include_statistics", false)
	if !includeStats && ctr.GetString("include_statistics", "") == "true" {
		includeStats = true
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_metric_info - metric=%s, include_statistics=%t - got called by session id: %s", metric, includeStats, sessionID)

	if metric == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metric_info failed after %v: missing required parameter: metric by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: metric")), nil
	}

	ret, err := k.GetPrometheusMetricInfo(metric, includeStats)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metric_info failed after %v: failed to get info for metric '%s': %v by session id: %s", duration, metric, err, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: Failed to get information for metric '%s': %v", metric, err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_metric_info completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusGenerateQuery handles the prometheus_generate_query tool request
func (s *Server) prometheusGenerateQuery(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_generate_query failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	description := ctr.GetString("description", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_generate_query - description=%s - got called by session id: %s", description, sessionID)

	if description == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_generate_query failed after %v: missing required parameter: description by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: description")), nil
	}

	// Generate the PromQL query
	query, err := k.GeneratePromQLQuery(description)
	if err != nil {
		duration := time.Since(start)
		errMsg := err.Error()
		if strings.Contains(errMsg, "failed to create LLM client") {
			klog.Errorf("Tool call: prometheus_generate_query failed after %v: could not connect to LLM service by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Could not connect to LLM service to generate PromQL query. The service may be unavailable.")), nil
		} else if strings.Contains(errMsg, "context deadline exceeded") || strings.Contains(errMsg, "timeout") {
			klog.Errorf("Tool call: prometheus_generate_query failed after %v: timeout occurred by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Timeout occurred while generating the PromQL query. Please try again with a simpler description or try later.")), nil
		}
		klog.Errorf("Tool call: prometheus_generate_query failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: Failed to generate PromQL query from description: %v", err)), nil
	}

	// Check if the response is empty or too short to be a valid query
	if len(query) < 5 {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_generate_query failed after %v: generated query too short: %s by session id: %s", duration, query, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: The generated query is too short or empty. Please provide a more specific description of the metric you're looking for.")), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_generate_query completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(query, nil), nil
}

// parseTime attempts to parse a time string in various formats
func parseTime(timeStr string, defaultTime time.Time) time.Time {
	// Try parsing as RFC3339
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return parsedTime
	}

	// Try parsing as Unix timestamp
	unixTime, err := strconv.ParseInt(timeStr, 10, 64)
	if err == nil {
		return time.Unix(unixTime, 0)
	}

	// Return default time if parsing fails
	return defaultTime
}

// prometheusSeries handles the prometheus_series_query tool request
func (s *Server) prometheusSeries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_series_query failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)

	// Extract the match parameter (required) using new API
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_series_query failed after %v: failed to get arguments by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	matchArg, ok := argsMap["match"]
	if !ok {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_series_query failed after %v: missing required parameter: match by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: match")), nil
	}

	// Convert the match parameter to a string slice
	matchSlice, ok := matchArg.([]interface{})
	if !ok {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_series_query failed after %v: match parameter must be a string array by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("match parameter must be a string array")), nil
	}

	// Convert the match slice to a string slice
	match := make([]string, len(matchSlice))
	for i, m := range matchSlice {
		match[i], ok = m.(string)
		if !ok {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_series_query failed after %v: match parameter must contain only strings by session id: %s", duration, sessionID)
			return NewTextResult("", errors.New("match parameter must contain only strings")), nil
		}
	}

	// Extract time window parameters - either start/end or time_window must be provided, default to 1h if neither provided
	timeWindowStr := ctr.GetString("time_window", "")
	var startTime, endTime *time.Time

	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_series_query failed after %v: invalid time_window format: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else {
		// Extract start parameter
		if startArg, exists := argsMap["start"]; exists && startArg != nil {
			if startStr, ok := startArg.(string); ok && startStr != "" {
				parsed := parseTime(startStr, time.Time{})
				if !parsed.IsZero() {
					startTime = &parsed
				} else {
					duration := time.Since(start)
					klog.Errorf("Tool call: prometheus_series_query failed after %v: invalid start format by session id: %s", duration, sessionID)
					return NewTextResult("", errors.New("invalid start format, use RFC3339 or Unix timestamp")), nil
				}
			}
		}

		// Extract end parameter
		if endArg, exists := argsMap["end"]; exists && endArg != nil {
			if endStr, ok := endArg.(string); ok && endStr != "" {
				parsed := parseTime(endStr, time.Time{})
				if !parsed.IsZero() {
					endTime = &parsed
				} else {
					duration := time.Since(start)
					klog.Errorf("Tool call: prometheus_series_query failed after %v: invalid end format by session id: %s", duration, sessionID)
					return NewTextResult("", errors.New("invalid end format, use RFC3339 or Unix timestamp")), nil
				}
			}
		}

		// If neither time_window nor start/end provided, default to 1h (reduced from 24h to prevent large responses)
		if startTime == nil && endTime == nil {
			now := time.Now()
			defaultStart := now.Add(-1 * time.Hour)
			startTime = &defaultStart
			endTime = &now
			klog.V(1).Infof("Tool call: prometheus_series_query - using default time_window of 1h by session id: %s", sessionID)
		} else if startTime == nil || endTime == nil {
			// If only one is provided, default the other to create 1h window
			now := time.Now()
			if startTime == nil {
				defaultStart := now.Add(-1 * time.Hour)
				startTime = &defaultStart
			}
			if endTime == nil {
				endTime = &now
			}
			klog.V(1).Infof("Tool call: prometheus_series_query - using default time_window of 1h (one time parameter missing) by session id: %s", sessionID)
		}

		// Validate start < end
		if startTime.After(*endTime) || startTime.Equal(*endTime) {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_series_query failed after %v: start time must be before end time by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("start time (%s) must be before end time (%s)", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))), nil
		}
	}

	// Extract optional limit parameter
	limit := 1000 // Default limit
	if limitArg, exists := argsMap["limit"]; exists && limitArg != nil {
		if limitVal, ok := limitArg.(float64); ok {
			limit = int(limitVal)
		}
	}

	var startStr, endStr string
	if startTime != nil {
		startStr = startTime.Format(time.RFC3339)
	}
	if endTime != nil {
		endStr = endTime.Format(time.RFC3339)
	}

	klog.V(1).Infof("Tool: prometheus_series_query - match_count=%d, start=%s, end=%s, limit=%d - got called by session id: %s",
		len(match), startStr, endStr, limit, sessionID)

	// Call the Kubernetes function
	ret, err := k.QueryPrometheusSeries(match, startTime, endTime, limit)
	if err != nil {
		duration := time.Since(start)
		errMsg := err.Error()
		// Check for common error patterns and provide more helpful messages
		if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
			klog.Errorf("Tool call: prometheus_series_query failed after %v: no series found matching selectors by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: No series found matching the provided selectors. The metrics may not exist in Prometheus or may have different labels than specified.")), nil
		} else if strings.Contains(errMsg, "parse error") || strings.Contains(errMsg, "bad_data") {
			klog.Errorf("Tool call: prometheus_series_query failed after %v: invalid series selector syntax by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Invalid series selector syntax in one of the match patterns: %v. Please check your selector format.", match)), nil
		} else if strings.Contains(errMsg, "failed to discover Prometheus") {
			klog.Errorf("Tool call: prometheus_series_query failed after %v: cannot connect to Prometheus server by session id: %s", duration, sessionID)
			return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
		}
		klog.Errorf("Tool call: prometheus_series_query failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("ERROR: Failed to query Prometheus series: %v", err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_series_query completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusTargets handles the prometheus_targets tool request
func (s *Server) prometheusTargets(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_targets failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	state := ctr.GetString("state", "")
	scrapePool := ctr.GetString("scrape_pool", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_targets - state=%s, scrape_pool=%s - got called by session id: %s", state, scrapePool, sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusTargets(state, scrapePool)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_targets failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Prometheus targets: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_targets completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusTargetMetadata handles the prometheus_targets_metadata tool request
func (s *Server) prometheusTargetMetadata(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_targets_metadata failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	matchTarget := ctr.GetString("match_target", "")
	metric := ctr.GetString("metric", "")

	// Extract optional limit parameter using new API
	limit := 0 // Default is no limit
	args := ctr.GetRawArguments()
	if argsMap, ok := args.(map[string]interface{}); ok {
		if limitArg, exists := argsMap["limit"]; exists && limitArg != nil {
			if limitVal, ok := limitArg.(float64); ok {
				limit = int(limitVal)
			}
		}
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_targets_metadata - match_target=%s, metric=%s, limit=%d - got called by session id: %s",
		matchTarget, metric, limit, sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusTargetMetadata(matchTarget, metric, limit)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_targets_metadata failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Prometheus target metadata: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_targets_metadata completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// Handler for creating Prometheus alerts
func (s *Server) prometheusCreateAlert(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_create_alert failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	alertName := ctr.GetString("alertname", "")
	expression := ctr.GetString("expression", "")
	appLabel := ctr.GetString("applabel", "")
	namespace := ctr.GetString("namespace", "")
	interval := ctr.GetString("interval", "")
	forDuration := ctr.GetString("for", "")

	args := ctr.GetRawArguments()
	var annotationsCount, alertLabelsCount int
	if argsMap, ok := args.(map[string]interface{}); ok {
		if annotations, exists := argsMap["annotations"]; exists {
			if annotationsMap, ok := annotations.(map[string]interface{}); ok {
				annotationsCount = len(annotationsMap)
			}
		}
		if alertLabels, exists := argsMap["alertlabels"]; exists {
			if alertLabelsMap, ok := alertLabels.(map[string]interface{}); ok {
				alertLabelsCount = len(alertLabelsMap)
			}
		}
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_create_alert - alertname=%s, applabel=%s, namespace=%s, interval=%s, for=%s, annotations_count=%d, alertlabels_count=%d - got called by session id: %s",
		alertName, appLabel, namespace, interval, forDuration, annotationsCount, alertLabelsCount, sessionID)

	// Extract and validate required parameters
	if alertName == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_create_alert failed after %v: missing required parameter: alertname by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: alertname")), nil
	}

	if expression == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_create_alert failed after %v: missing required parameter: expression by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: expression")), nil
	}

	if appLabel == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_create_alert failed after %v: missing required parameter: applabel by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: applabel")), nil
	}

	if namespace == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_create_alert failed after %v: missing required parameter: namespace by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: namespace")), nil
	}

	// Extract optional parameters with defaults
	if interval == "" {
		interval = "1m" // Default to 1 minute
	}

	if forDuration == "" {
		forDuration = "5m" // Default to 5 minutes
	}

	// Convert annotations from interface{} to map[string]string using new API
	var annotations map[string]string
	if argsMap, ok := args.(map[string]interface{}); ok {
		if annotationsRaw, exists := argsMap["annotations"]; exists && annotationsRaw != nil {
			annotations = make(map[string]string)
			if annotationsMap, ok := annotationsRaw.(map[string]interface{}); ok {
				for k, v := range annotationsMap {
					if str, ok := v.(string); ok {
						annotations[k] = str
					}
				}
			}
		}

		// Convert alertlabels from interface{} to map[string]string using new API
		var alertLabels map[string]string
		if alertLabelsRaw, exists := argsMap["alertlabels"]; exists && alertLabelsRaw != nil {
			alertLabels = make(map[string]string)
			if alertLabelsMap, ok := alertLabelsRaw.(map[string]interface{}); ok {
				for k, v := range alertLabelsMap {
					if str, ok := v.(string); ok {
						alertLabels[k] = str
					}
				}
			}
		}

		// Call the Kubernetes function
		result, err := k.CreatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration, annotations, alertLabels)
		if err != nil {
			duration := time.Since(start)
			klog.Errorf("Tool call: prometheus_create_alert failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to create Prometheus alert: %v", err)), nil
		}

		duration := time.Since(start)
		klog.V(1).Infof("Tool call: prometheus_create_alert completed successfully in %v by session id: %s", duration, sessionID)
		return NewTextResult(result, nil), nil
	}

	duration := time.Since(start)
	klog.Errorf("Tool call: prometheus_create_alert failed after %v: failed to get arguments by session id: %s", duration, sessionID)
	return NewTextResult("", errors.New("failed to get arguments")), nil
}

// Handler for updating Prometheus alerts
func (s *Server) prometheusUpdateAlert(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_update_alert failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	alertName := ctr.GetString("alertname", "")
	appLabel := ctr.GetString("applabel", "")
	namespace := ctr.GetString("namespace", "")
	expression := ctr.GetString("expression", "")
	interval := ctr.GetString("interval", "")
	forDuration := ctr.GetString("for", "")

	args := ctr.GetRawArguments()
	var annotationsCount, alertLabelsCount int
	if argsMap, ok := args.(map[string]interface{}); ok {
		if annotations, exists := argsMap["annotations"]; exists {
			if annotationsMap, ok := annotations.(map[string]interface{}); ok {
				annotationsCount = len(annotationsMap)
			}
		}
		if alertLabels, exists := argsMap["alertlabels"]; exists {
			if alertLabelsMap, ok := alertLabels.(map[string]interface{}); ok {
				alertLabelsCount = len(alertLabelsMap)
			}
		}
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_update_alert - alertname=%s, applabel=%s, namespace=%s, expression=%s, interval=%s, for=%s, annotations_count=%d, alertlabels_count=%d - got called by session id: %s",
		alertName, appLabel, namespace, expression, interval, forDuration, annotationsCount, alertLabelsCount, sessionID)

	// Extract and validate required parameters
	if alertName == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_update_alert failed after %v: missing required parameter: alertname by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: alertname")), nil
	}

	if appLabel == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_update_alert failed after %v: missing required parameter: applabel by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: applabel")), nil
	}

	if namespace == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_update_alert failed after %v: missing required parameter: namespace by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: namespace")), nil
	}

	// Convert annotations and alertlabels using new API
	var annotations map[string]string
	var alertLabels map[string]string

	if argsMap, ok := args.(map[string]interface{}); ok {
		// Convert annotations from interface{} to map[string]string
		if annotationsRaw, exists := argsMap["annotations"]; exists && annotationsRaw != nil {
			annotations = make(map[string]string)
			if annotationsMap, ok := annotationsRaw.(map[string]interface{}); ok {
				for k, v := range annotationsMap {
					if str, ok := v.(string); ok {
						annotations[k] = str
					}
				}
			}
		}

		// Convert alertlabels from interface{} to map[string]string
		if alertLabelsRaw, exists := argsMap["alertlabels"]; exists && alertLabelsRaw != nil {
			alertLabels = make(map[string]string)
			if alertLabelsMap, ok := alertLabelsRaw.(map[string]interface{}); ok {
				for k, v := range alertLabelsMap {
					if str, ok := v.(string); ok {
						alertLabels[k] = str
					}
				}
			}
		}
	}

	// Call the Kubernetes function (remove type casting since these are already strings)
	result, err := k.UpdatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration, annotations, alertLabels)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_update_alert failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to update Prometheus alert: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_update_alert completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(result, nil), nil
}

// Handler for deleting Prometheus alerts
func (s *Server) prometheusDeleteAlert(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_delete_alert failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	appLabel := ctr.GetString("applabel", "")
	namespace := ctr.GetString("namespace", "")
	alertName := ctr.GetString("alertname", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_delete_alert - applabel=%s, namespace=%s, alertname=%s - got called by session id: %s",
		appLabel, namespace, alertName, sessionID)

	// Extract and validate required parameters
	if appLabel == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_delete_alert failed after %v: missing required parameter: applabel by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: applabel")), nil
	}

	if namespace == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_delete_alert failed after %v: missing required parameter: namespace by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: namespace")), nil
	}

	// Call the Kubernetes function
	result, err := k.DeletePrometheusAlert(appLabel, namespace, alertName)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_delete_alert failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to delete Prometheus alert: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_delete_alert completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(result, nil), nil
}

// prometheusGetAlerts handles the prometheus_get_alerts tool request
func (s *Server) prometheusGetAlerts(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_get_alerts failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: prometheus_get_alerts failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: prometheus_get_alerts failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: prometheus_get_alerts failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: prometheus_get_alerts failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed

		// Validate start < end
		if startTime.After(*endTime) || startTime.Equal(*endTime) {
			klog.Errorf("Tool call: prometheus_get_alerts failed after %v: start_time must be before end_time", time.Since(start))
			return NewTextResult("", fmt.Errorf("start_time (%s) must be before end_time (%s)", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))), nil
		}
	} else {
		// Neither time_window nor start_time/end_time provided, default to 1h (reduced from 24h to prevent large responses)
		now := time.Now()
		defaultStart := now.Add(-1 * time.Hour)
		startTime = &defaultStart
		endTime = &now
		klog.V(1).Infof("Tool call: prometheus_get_alerts - using default time_window of 1h by session id: %s", getSessionID(ctx))
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_get_alerts - got called by session id: %s", sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusAlerts(startTime, endTime)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_get_alerts failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Prometheus alerts: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_get_alerts completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusGetRules handles the prometheus_get_rules tool request
func (s *Server) prometheusGetRules(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_get_rules failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	// Extract optional parameters using new API
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_get_rules failed after %v: failed to get arguments", duration)
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	// Extract rule_name parameter (string array)
	var ruleNames []string
	if ruleNamesArg, exists := argsMap["rule_name"]; exists && ruleNamesArg != nil {
		if ruleNamesArr, ok := ruleNamesArg.([]interface{}); ok {
			for _, name := range ruleNamesArr {
				if nameStr, ok := name.(string); ok {
					ruleNames = append(ruleNames, nameStr)
				}
			}
		}
	}

	// Extract rule_group parameter (string array)
	var ruleGroups []string
	if ruleGroupsArg, exists := argsMap["rule_group"]; exists && ruleGroupsArg != nil {
		if ruleGroupsArr, ok := ruleGroupsArg.([]interface{}); ok {
			for _, group := range ruleGroupsArr {
				if groupStr, ok := group.(string); ok {
					ruleGroups = append(ruleGroups, groupStr)
				}
			}
		}
	}

	// Extract file parameter (string array)
	var files []string
	if filesArg, exists := argsMap["file"]; exists && filesArg != nil {
		if filesArr, ok := filesArg.([]interface{}); ok {
			for _, file := range filesArr {
				if fileStr, ok := file.(string); ok {
					files = append(files, fileStr)
				}
			}
		}
	}

	// Extract exclude_alerts parameter (boolean) using GetBool
	excludeAlerts := ctr.GetBool("exclude_alerts", false)

	// Extract match parameter (string array)
	var matchLabels []string
	if matchArg, exists := argsMap["match"]; exists && matchArg != nil {
		if matchArr, ok := matchArg.([]interface{}); ok {
			for _, match := range matchArr {
				if matchStr, ok := match.(string); ok {
					matchLabels = append(matchLabels, matchStr)
				}
			}
		}
	}

	// Extract group_limit parameter (number) using GetFloat
	groupLimit := int(ctr.GetFloat("group_limit", 0))

	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: prometheus_get_rules failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: prometheus_get_rules failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: prometheus_get_rules failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: prometheus_get_rules failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed

		// Validate start < end
		if startTime.After(*endTime) || startTime.Equal(*endTime) {
			klog.Errorf("Tool call: prometheus_get_rules failed after %v: start_time must be before end_time", time.Since(start))
			return NewTextResult("", fmt.Errorf("start_time (%s) must be before end_time (%s)", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))), nil
		}
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_get_rules - rule_names_count=%d, rule_groups_count=%d, files_count=%d, exclude_alerts=%t, match_labels_count=%d, group_limit=%d - got called by session id: %s",
		len(ruleNames), len(ruleGroups), len(files), excludeAlerts, len(matchLabels), groupLimit, sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusRules(groupLimit, ruleNames, ruleGroups, files, excludeAlerts, matchLabels, startTime, endTime)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_get_rules failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Prometheus rules: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_get_rules completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusRuntimeInfo handles the prometheus_runtimeinfo tool request
func (s *Server) prometheusRuntimeInfo(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_runtimeinfo failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_runtimeinfo - got called by session id: %s", sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusRuntimeInfo()
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_runtimeinfo failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Prometheus runtime info: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_runtimeinfo completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusTSDBStatus handles the prometheus_TSDB_status tool request
func (s *Server) prometheusTSDBStatus(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_TSDB_status failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	limit := int(ctr.GetFloat("limit", 0)) // Default is no limit

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_TSDB_status - limit=%d - got called by session id: %s", limit, sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusTSDBStatus(limit)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_TSDB_status failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Prometheus TSDB status: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_TSDB_status completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusListLabelNames handles the prometheus_list_label_names tool request
func (s *Server) prometheusListLabelNames(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_list_label_names failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	startRfc3339 := ctr.GetString("start", ctr.GetString("startRfc3339", ""))
	endRfc3339 := ctr.GetString("end", ctr.GetString("endRfc3339", ""))
	limit := int(ctr.GetFloat("limit", 0))

	// Extract matches parameter using GetRawArguments
	var matches []string
	args := ctr.GetRawArguments()
	if argsMap, ok := args.(map[string]interface{}); ok {
		if matchesArg, exists := argsMap["matches"]; exists && matchesArg != nil {
			if matchesArr, ok := matchesArg.([]interface{}); ok {
				for _, m := range matchesArr {
					if matchStr, ok := m.(string); ok {
						matches = append(matches, matchStr)
					}
				}
			}
		}
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_list_label_names - start=%s, end=%s, limit=%d, matches_count=%d - got called by session id: %s",
		startRfc3339, endRfc3339, limit, len(matches), sessionID)

	// Call the Kubernetes function
	ret, err := k.ListPrometheusLabelNames(startRfc3339, endRfc3339, limit, matches)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_list_label_names failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list Prometheus label names: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_list_label_names completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// prometheusListLabelValues handles the prometheus_list_label_values tool request
func (s *Server) prometheusListLabelValues(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: prometheus_list_label_values failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	labelName := ctr.GetString("labelName", "")
	startRfc3339 := ctr.GetString("start", ctr.GetString("startRfc3339", ""))
	endRfc3339 := ctr.GetString("end", ctr.GetString("endRfc3339", ""))
	limit := int(ctr.GetFloat("limit", 0))

	// Extract matches parameter using GetRawArguments
	var matches []string
	args := ctr.GetRawArguments()
	if argsMap, ok := args.(map[string]interface{}); ok {
		if matchesArg, exists := argsMap["matches"]; exists && matchesArg != nil {
			if matchesArr, ok := matchesArg.([]interface{}); ok {
				for _, m := range matchesArr {
					if matchStr, ok := m.(string); ok {
						matches = append(matches, matchStr)
					}
				}
			}
		}
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_list_label_values - labelName=%s, start=%s, end=%s, limit=%d, matches_count=%d - got called by session id: %s",
		labelName, startRfc3339, endRfc3339, limit, len(matches), sessionID)

	// Extract required parameters
	if labelName == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_list_label_values failed after %v: missing required parameter: labelName by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: labelName")), nil
	}

	// Call the Kubernetes function
	ret, err := k.ListPrometheusLabelValues(labelName, startRfc3339, endRfc3339, limit, matches)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_list_label_values failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to list Prometheus label values: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prometheus_list_label_values completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}
