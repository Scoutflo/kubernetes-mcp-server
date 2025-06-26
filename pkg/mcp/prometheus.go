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
		{Tool: mcp.NewTool("prometheus_generate_query",
			mcp.WithDescription("Tool for generating a PromQL query from a natural language description. Use this tool when you need to create a PromQL expression based on a natural language description of the metric you want to query."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("description", mcp.Description("Natural language description of the metric you want to query"), mcp.Required()),
		), Handler: s.prometheusGenerateQuery},
		{Tool: mcp.NewTool("prometheus_metrics_query",
			mcp.WithDescription("Tool for executing instant Prometheus queries. Executes instant queries against Prometheus to retrieve current metric values. Use this tool when you need to get the latest values of metrics or perform calculations on current data. The query must be a valid PromQL expression."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("query", mcp.Description("Prometheus PromQL expression query string"), mcp.Required()),
			mcp.WithString("time", mcp.Description("Evaluation timestamp in RFC3339 or unix timestamp format (optional)")),
			mcp.WithString("timeout", mcp.Description("Evaluation timeout (optional)")),
		), Handler: s.prometheusMetrics},
		{Tool: mcp.NewTool("prometheus_metrics_query_range",
			mcp.WithDescription("Tool for executing range queries in Prometheus. Executes time series queries over a specified time range in Prometheus. Use this tool for analyzing metric patterns, trends, and historical data. You can specify the time range, resolution (step), and timeout for the query."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("query", mcp.Description("Prometheus PromQL expression query string"), mcp.Required()),
			mcp.WithString("start", mcp.Description("Start timestamp in RFC3339 or Unix timestamp format"), mcp.Required()),
			mcp.WithString("end", mcp.Description("End timestamp in RFC3339 or Unix timestamp format"), mcp.Required()),
			mcp.WithString("step", mcp.Description("Query resolution step width (e.g., '15s', '1m', '1h')"), mcp.Required()),
			mcp.WithString("timeout", mcp.Description("Evaluation timeout (optional)")),
		), Handler: s.prometheusMetricsRange},
		{Tool: mcp.NewTool("prometheus_list_metrics",
			mcp.WithDescription("Tool for listing all available Prometheus metrics with their metadata. Use this tool to discover all metrics available in Prometheus and their associated information."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.prometheusListMetrics},
		{Tool: mcp.NewTool("prometheus_metric_info",
			mcp.WithDescription("Tool for getting detailed information about a specific Prometheus metric. Use this tool to discover more about a particular metric and its associated statistics. You can include count, min, max, and avg statistics for the metric if needed."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("metric", mcp.Description("Name of the metric to get information about"), mcp.Required()),
			mcp.WithBoolean("include_statistics", mcp.Description("Include count, min, max, and avg statistics for this metric. May be slower for metrics with many time series.")),
		), Handler: s.prometheusMetricInfo},
		{Tool: mcp.NewTool("prometheus_series_query",
			mcp.WithDescription("Tool for querying Prometheus series. Finds time series that match certain label selectors in Prometheus. Use this tool to discover which metrics exist and their label combinations. You can specify time ranges to limit the search scope and set a maximum number of results."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
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
			mcp.WithNumber("limit", mcp.Description("Maximum number of returned items (optional)")),
		), Handler: s.prometheusSeries},
		{Tool: mcp.NewTool("prometheus_targets",
			mcp.WithDescription("Tool for getting Prometheus target discovery state. Provides information about all Prometheus scrape targets and their current state. Use this tool to monitor which targets are being scraped successfully and which are failing. You can filter targets by state (active/dropped) and scrape pool."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("state", mcp.Description("Target state filter, must be one of: active, dropped, any (optional)")),
			mcp.WithString("scrape_pool", mcp.Description("Scrape pool name (optional)")),
		), Handler: s.prometheusTargets},
		{Tool: mcp.NewTool("prometheus_targets_metadata",
			mcp.WithDescription("Tool for getting Prometheus target metadata. Retrieves metadata about metrics exposed by specific Prometheus targets. Use this tool to understand metric types, help texts, and units. You can filter by target labels and specific metric names."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("match_target", mcp.Description("Target label selectors (optional)")),
			mcp.WithString("metric", mcp.Description("Metric name (optional)")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of targets (optional)")),
		), Handler: s.prometheusTargetMetadata},
		{Tool: mcp.NewTool("prometheus_list_label_names",
			mcp.WithDescription("List label names in a Prometheus datasource. Allows filtering by series selectors and time range."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("startRfc3339", mcp.Description("Optionally, the start time of the time range to filter the results by")),
			mcp.WithString("endRfc3339", mcp.Description("Optionally, the end time of the time range to filter the results by")),
			mcp.WithNumber("limit", mcp.Description("Optionally, the maximum number of results to return")),
			mcp.WithArray("matches", mcp.Description("Optionally, a list of label matchers to filter the results by"),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
		), Handler: s.prometheusListLabelNames},
		{Tool: mcp.NewTool("prometheus_list_label_values",
			mcp.WithDescription("Get the values for a specific label name in Prometheus. Allows filtering by series selectors and time range."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("labelName", mcp.Description("The name of the label to query"), mcp.Required()),
			mcp.WithString("startRfc3339", mcp.Description("Optionally, the start time of the query")),
			mcp.WithString("endRfc3339", mcp.Description("Optionally, the end time of the query")),
			mcp.WithNumber("limit", mcp.Description("Optionally, the maximum number of results to return")),
			mcp.WithArray("matches", mcp.Description("Optionally, a list of selectors to filter the results by"),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
		), Handler: s.prometheusListLabelValues},
		{Tool: mcp.NewTool("prometheus_get_alerts",
			mcp.WithDescription("Tool for getting active Prometheus alerts. Retrieves all currently firing alerts in the Prometheus server. Use this tool to monitor the current alert state and identify ongoing issues. Returns details about alert names, labels, and when they started firing."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.prometheusGetAlerts},
		{Tool: mcp.NewTool("prometheus_get_rules",
			mcp.WithDescription("Tool for getting Prometheus alerting and recording rules. Retrieves information about configured alerting and recording rules in Prometheus. Use this tool to understand what alerts are defined and what metrics are being pre-computed. You can filter rules by type, name, group, and other criteria."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
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
			mcp.WithBoolean("exclude_alerts", mcp.Description("Exclude alerts flag")),
			mcp.WithArray("match", mcp.Description("Label selectors"),
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
			mcp.WithNumber("group_limit", mcp.Description("Group limit")),
		), Handler: s.prometheusGetRules},
		{Tool: mcp.NewTool("prometheus_create_alert",
			mcp.WithDescription("Tool for creating a new Prometheus alert rule. This tool allows you to define alerting rules that will trigger notifications when specific conditions are met. You can customize the alert with annotations, labels, and evaluation intervals."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("alertname", mcp.Description("Name of the alert to create"), mcp.Required()),
			mcp.WithString("expression", mcp.Description("PromQL expression that defines the alert condition, If not provided, please generate a query using prometheus_generate_query tool"), mcp.Required()),
			mcp.WithString("applabel", mcp.Description("Application label used to identify the PrometheusRule resource, use alertname if applabel is not provided"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Kubernetes namespace to create the alert in"), mcp.Required()),
			mcp.WithString("interval", mcp.Description("Evaluation interval for the alert group (e.g., '30s', '1m', '5m')")),
			mcp.WithString("for", mcp.Description("Duration for which the condition must be true before firing (e.g., '5m')")),
			mcp.WithObject("annotations", mcp.Description("Map of annotations to add to the alert (description, summary, etc.)")),
			mcp.WithObject("alertlabels", mcp.Description("Map of labels to attach to the alert")),
		), Handler: s.prometheusCreateAlert},
		{Tool: mcp.NewTool("prometheus_update_alert",
			mcp.WithDescription("Tool for updating an existing Prometheus alert rule. This tool allows you to modify the properties of an existing alert without having to delete and recreate it. You can update the condition, annotations, labels, and other attributes."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("alertname", mcp.Description("Name of the alert to update"), mcp.Required()),
			mcp.WithString("applabel", mcp.Description("Application label that identifies the PrometheusRule resource, use alertname if applabel is not provided"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Kubernetes namespace of the alert"), mcp.Required()),
			mcp.WithString("expression", mcp.Description("New PromQL expression for the alert condition")),
			mcp.WithString("interval", mcp.Description("New evaluation interval for the alert group (e.g., '30s', '1m', '5m')")),
			mcp.WithString("for", mcp.Description("New duration for which the condition must be true before firing (e.g., '5m')")),
			mcp.WithObject("annotations", mcp.Description("New or updated annotations for the alert")),
			mcp.WithObject("alertlabels", mcp.Description("New or updated labels for the alert")),
		), Handler: s.prometheusUpdateAlert},
		{Tool: mcp.NewTool("prometheus_delete_alert",
			mcp.WithDescription("Tool for deleting a Prometheus alert rule. This tool removes an existing alert rule from the system. You can either delete a specific alert within a rule group or the entire PrometheusRule resource."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithString("applabel", mcp.Description("Application label that identifies the PrometheusRule resource, use alertname if applabel is not provided"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Kubernetes namespace of the alert"), mcp.Required()),
			mcp.WithString("alertname", mcp.Description("Name of the specific alert to delete within the rule group (optional)")),
		), Handler: s.prometheusDeleteAlert},
		{Tool: mcp.NewTool("prometheus_runtimeinfo",
			mcp.WithDescription("Tool for getting Prometheus runtime information. Provides detailed information about the Prometheus server's runtime state. Use this tool to monitor server health and performance through details about garbage collection, memory usage, and other runtime metrics."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.prometheusRuntimeInfo},
		{Tool: mcp.NewTool("prometheus_TSDB_status",
			mcp.WithDescription("Tool for getting Prometheus TSDB status. Provides information about the time series database (TSDB) status in Prometheus. Use this tool to monitor database health through details about data storage, head blocks, WAL status, and other TSDB metrics."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
			mcp.WithNumber("limit", mcp.Description("Number of items limit")),
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
	timeout := ctr.GetString("timeout", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool call: prometheus_metrics_query - query=%s, time=%s, timeout=%s - got called by session id: %s", queryArg, timeArg, timeout, sessionID)

	if queryArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query failed after %v: missing required parameter: query by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	query := queryArg

	// Extract optional time parameter
	var queryTime *time.Time
	if timeArg != "" {
		parsedTime := parseTime(timeArg, time.Time{})
		if !parsedTime.IsZero() {
			queryTime = &parsedTime
		}
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
	timeout := ctr.GetString("timeout", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool call: prometheus_metrics_query_range - query=%s, start=%s, end=%s, step=%s, timeout=%s - got called by session id: %s",
		queryArg, startArg, endArg, stepArg, timeout, sessionID)

	// Validate required parameters
	if queryArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: query by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	if startArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: start by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: start")), nil
	}
	if endArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: end by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: end")), nil
	}
	if stepArg == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metrics_query_range failed after %v: missing required parameter: step by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: step")), nil
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

	// Parse step
	step := stepArg

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
	includeStatsArg := ctr.GetString("include_statistics", "")

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_metric_info - metric=%s, include_statistics=%s - got called by session id: %s", metric, includeStatsArg, sessionID)

	if metric == "" {
		duration := time.Since(start)
		klog.Errorf("Tool call: prometheus_metric_info failed after %v: missing required parameter: metric by session id: %s", duration, sessionID)
		return NewTextResult("", errors.New("missing required parameter: metric")), nil
	}

	// Check if statistics are requested
	includeStats := includeStatsArg == "true"

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

	// Extract optional start parameter
	var startTime *time.Time
	if startArg, exists := argsMap["start"]; exists && startArg != nil {
		if startStr, ok := startArg.(string); ok {
			parsed := parseTime(startStr, time.Time{})
			if !parsed.IsZero() {
				startTime = &parsed
			}
		}
	}

	// Extract optional end parameter
	var endTime *time.Time
	if endArg, exists := argsMap["end"]; exists && endArg != nil {
		if endStr, ok := endArg.(string); ok {
			parsed := parseTime(endStr, time.Time{})
			if !parsed.IsZero() {
				endTime = &parsed
			}
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
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_get_alerts - got called by session id: %s", sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusAlerts()
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
	groupLimit := ""
	if limitVal := ctr.GetFloat("group_limit", 0); limitVal > 0 {
		groupLimit = fmt.Sprintf("%d", int(limitVal))
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prometheus_get_rules - rule_names_count=%d, rule_groups_count=%d, files_count=%d, exclude_alerts=%t, match_labels_count=%d, group_limit=%s - got called by session id: %s",
		len(ruleNames), len(ruleGroups), len(files), excludeAlerts, len(matchLabels), groupLimit, sessionID)

	// Call the Kubernetes function
	ret, err := k.GetPrometheusRules(groupLimit, ruleNames, ruleGroups, files, excludeAlerts, matchLabels)
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
	startRfc3339 := ctr.GetString("startRfc3339", "")
	endRfc3339 := ctr.GetString("endRfc3339", "")
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
	klog.V(1).Infof("Tool: prometheus_list_label_names - startRfc3339=%s, endRfc3339=%s, limit=%d, matches_count=%d - got called by session id: %s",
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
	startRfc3339 := ctr.GetString("startRfc3339", "")
	endRfc3339 := ctr.GetString("endRfc3339", "")
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
	klog.V(1).Infof("Tool: prometheus_list_label_values - labelName=%s, startRfc3339=%s, endRfc3339=%s, limit=%d, matches_count=%d - got called by session id: %s",
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
