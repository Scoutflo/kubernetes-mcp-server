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
)

func (s *Server) initPrometheus() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("prometheus_generate_query",
			mcp.WithDescription("Tool for generating a PromQL query from a natural language description. Use this tool when you need to create a PromQL expression based on a natural language description of the metric you want to query."),
			mcp.WithString("description", mcp.Description("Natural language description of the metric you want to query"), mcp.Required()),
		), Handler: s.prometheusGenerateQuery},
		{Tool: mcp.NewTool("prometheus_metrics_query",
			mcp.WithDescription("Tool for executing instant Prometheus queries. Executes instant queries against Prometheus to retrieve current metric values. Use this tool when you need to get the latest values of metrics or perform calculations on current data. The query must be a valid PromQL expression."),
			mcp.WithString("query", mcp.Description("Prometheus PromQL expression query string"), mcp.Required()),
			mcp.WithString("time", mcp.Description("Evaluation timestamp in RFC3339 or unix timestamp format (optional)")),
			mcp.WithString("timeout", mcp.Description("Evaluation timeout (optional)")),
		), Handler: s.prometheusMetrics},
		{Tool: mcp.NewTool("prometheus_metrics_query_range",
			mcp.WithDescription("Tool for executing range queries in Prometheus. Executes time series queries over a specified time range in Prometheus. Use this tool for analyzing metric patterns, trends, and historical data. You can specify the time range, resolution (step), and timeout for the query."),
			mcp.WithString("query", mcp.Description("Prometheus PromQL expression query string"), mcp.Required()),
			mcp.WithString("start", mcp.Description("Start timestamp in RFC3339 or Unix timestamp format"), mcp.Required()),
			mcp.WithString("end", mcp.Description("End timestamp in RFC3339 or Unix timestamp format"), mcp.Required()),
			mcp.WithString("step", mcp.Description("Query resolution step width (e.g., '15s', '1m', '1h')"), mcp.Required()),
			mcp.WithString("timeout", mcp.Description("Evaluation timeout (optional)")),
		), Handler: s.prometheusMetricsRange},
		{Tool: mcp.NewTool("prometheus_list_metrics",
			mcp.WithDescription("Tool for listing all available Prometheus metrics with their metadata. Use this tool to discover all metrics available in Prometheus and their associated information."),
		), Handler: s.prometheusListMetrics},
		{Tool: mcp.NewTool("prometheus_metric_info",
			mcp.WithDescription("Tool for getting detailed information about a specific Prometheus metric. Use this tool to discover more about a particular metric and its associated statistics. You can include count, min, max, and avg statistics for the metric if needed."),
			mcp.WithString("metric", mcp.Description("Name of the metric to get information about"), mcp.Required()),
			mcp.WithBoolean("include_statistics", mcp.Description("Include count, min, max, and avg statistics for this metric. May be slower for metrics with many time series.")),
		), Handler: s.prometheusMetricInfo},
		{Tool: mcp.NewTool("prometheus_series_query",
			mcp.WithDescription("Tool for querying Prometheus series. Finds time series that match certain label selectors in Prometheus. Use this tool to discover which metrics exist and their label combinations. You can specify time ranges to limit the search scope and set a maximum number of results."),
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
			mcp.WithString("state", mcp.Description("Target state filter, must be one of: active, dropped, any (optional)")),
			mcp.WithString("scrape_pool", mcp.Description("Scrape pool name (optional)")),
		), Handler: s.prometheusTargets},
		{Tool: mcp.NewTool("prometheus_targets_metadata",
			mcp.WithDescription("Tool for getting Prometheus target metadata. Retrieves metadata about metrics exposed by specific Prometheus targets. Use this tool to understand metric types, help texts, and units. You can filter by target labels and specific metric names."),
			mcp.WithString("match_target", mcp.Description("Target label selectors (optional)")),
			mcp.WithString("metric", mcp.Description("Metric name (optional)")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of targets (optional)")),
		), Handler: s.prometheusTargetMetadata},
		{Tool: mcp.NewTool("prometheus_get_alerts",
			mcp.WithDescription("Tool for getting active Prometheus alerts. Retrieves all currently firing alerts in the Prometheus server. Use this tool to monitor the current alert state and identify ongoing issues. Returns details about alert names, labels, and when they started firing."),
		), Handler: s.prometheusGetAlerts},
		{Tool: mcp.NewTool("prometheus_get_rules",
			mcp.WithDescription("Tool for getting Prometheus alerting and recording rules. Retrieves information about configured alerting and recording rules in Prometheus. Use this tool to understand what alerts are defined and what metrics are being pre-computed. You can filter rules by type, name, group, and other criteria."),
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
			mcp.WithString("applabel", mcp.Description("Application label that identifies the PrometheusRule resource, use alertname if applabel is not provided"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Kubernetes namespace of the alert"), mcp.Required()),
			mcp.WithString("alertname", mcp.Description("Name of the specific alert to delete within the rule group (optional)")),
		), Handler: s.prometheusDeleteAlert},
		// {Tool: mcp.NewTool("prometheus_clean_tombstones",
		// 	mcp.WithDescription("Tool for cleaning Prometheus tombstones. Removes tombstone files created during Prometheus data deletion operations. Use this tool to maintain database cleanliness and recover storage space. Tombstones are markers for deleted data and can be safely removed after their retention period."),
		// ), Handler: s.prometheusCleanTombstones},
		// {Tool: mcp.NewTool("prometheus_create_snapshot",
		// 	mcp.WithDescription("Tool for creating Prometheus snapshots. Creates a snapshot of the current Prometheus TSDB data. Use this tool for backup purposes or creating point-in-time copies of the data. You can optionally skip snapshotting the head block (latest, incomplete data)."),
		// 	mcp.WithBoolean("skip_head", mcp.Description("Skip head block flag")),
		// ), Handler: s.prometheusCreateSnapshot},
		// {Tool: mcp.NewTool("prometheus_delete_series",
		// 	mcp.WithDescription("Tool for deleting Prometheus series data. Deletes time series data matching specific criteria in Prometheus. Use this tool carefully to remove obsolete data or free up storage space. Deleted data cannot be recovered. You can specify time ranges and series selectors."),
		// 	mcp.WithArray("match", mcp.Description("Series selectors"),
		// 		func(schema map[string]interface{}) {
		// 			schema["type"] = "array"
		// 			schema["items"] = map[string]interface{}{
		// 				"type": "string",
		// 			}
		// 		},
		// 		mcp.Required()),
		// 	mcp.WithString("start", mcp.Description("Start timestamp in RFC3339 or Unix timestamp format (optional)")),
		// 	mcp.WithString("end", mcp.Description("End timestamp in RFC3339 or Unix timestamp format (optional)")),
		// ), Handler: s.prometheusDeleteSeries},
		// {Tool: mcp.NewTool("prometheus_alert_manager",
		// 	mcp.WithDescription("Tool for getting Prometheus alertmanager discovery state. Provides information about the Alertmanager instances known to Prometheus. Use this tool to verify the connection status between Prometheus and its Alertmanagers. Shows both active and dropped Alertmanager instances."),
		// ), Handler: s.prometheusAlertManagers},
		{Tool: mcp.NewTool("prometheus_runtimeinfo",
			mcp.WithDescription("Tool for getting Prometheus runtime information. Provides detailed information about the Prometheus server's runtime state. Use this tool to monitor server health and performance through details about garbage collection, memory usage, and other runtime metrics."),
		), Handler: s.prometheusRuntimeInfo},
		{Tool: mcp.NewTool("prometheus_TSDB_status",
			mcp.WithDescription("Tool for getting Prometheus TSDB status. Provides information about the time series database (TSDB) status in Prometheus. Use this tool to monitor database health through details about data storage, head blocks, WAL status, and other TSDB metrics."),
			mcp.WithNumber("limit", mcp.Description("Number of items limit")),
		), Handler: s.prometheusTSDBStatus},
		// {Tool: mcp.NewTool("prometheus_WALReplay",
		// 	mcp.WithDescription("Tool for getting Prometheus WAL replay status. Retrieves the status of Write-Ahead Log (WAL) replay operations in Prometheus. Use this tool to monitor the progress of WAL replay during server startup or recovery. Helps track data durability and recovery progress."),
		// ), Handler: s.prometheusWALReplay},
	}
}

// prometheusMetrics handles the prometheus_metrics_query tool request
func (s *Server) prometheusMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required query parameter
	queryArg := ctr.GetString("query", "")
	if queryArg == "" {
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	query := queryArg

	// Extract optional time parameter
	var queryTime *time.Time
	if timeArg := ctr.GetString("time", ""); timeArg != "" {
		parsedTime := parseTime(timeArg, time.Time{})
		if !parsedTime.IsZero() {
			queryTime = &parsedTime
		}
	}

	// Extract optional timeout parameter
	timeout := ""
	if timeoutArg := ctr.GetString("timeout", ""); timeoutArg != "" {
		timeout = timeoutArg
	}

	// Execute the instant query with the provided parameters
	ret, err := s.k.QueryPrometheus(query, queryTime, timeout)
	if err != nil {
		errMsg := err.Error()
		// Check for common error patterns and provide more helpful messages
		if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
			return NewTextResult("", fmt.Errorf("ERROR: Metric not found. The specified metric '%s' does not exist in Prometheus. Please check the metric name and ensure it's correctly spelled.", query)), nil
		} else if strings.Contains(errMsg, "parse error") {
			return NewTextResult("", fmt.Errorf("ERROR: Invalid PromQL query syntax in '%s'. Please check your query format.", query)), nil
		} else if strings.Contains(errMsg, "failed to discover Prometheus") {
			return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
		}
		// For other errors, return a clear error message
		return NewTextResult("", fmt.Errorf("ERROR: Failed to execute Prometheus query: %v", err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	return NewTextResult(ret, nil), nil
}

// prometheusMetricsRange handles the prometheus_metrics_query_range tool request
func (s *Server) prometheusMetricsRange(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	queryArg := ctr.GetString("query", "")
	startArg := ctr.GetString("start", "")
	endArg := ctr.GetString("end", "")
	stepArg := ctr.GetString("step", "")

	// Validate required parameters
	if queryArg == "" {
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	if startArg == "" {
		return NewTextResult("", errors.New("missing required parameter: start")), nil
	}
	if endArg == "" {
		return NewTextResult("", errors.New("missing required parameter: end")), nil
	}
	if stepArg == "" {
		return NewTextResult("", errors.New("missing required parameter: step")), nil
	}

	// Parse query
	query := queryArg

	// Parse start time
	startTime := parseTime(startArg, time.Now().Add(-1*time.Hour))
	if startTime.IsZero() {
		return NewTextResult("", errors.New("invalid start time format")), nil
	}

	// Parse end time
	endTime := parseTime(endArg, time.Now())
	if endTime.IsZero() {
		return NewTextResult("", errors.New("invalid end time format")), nil
	}

	// Parse step
	step := stepArg

	// Extract optional timeout parameter
	timeout := ""
	if timeoutArg := ctr.GetString("timeout", ""); timeoutArg != "" {
		timeout = timeoutArg
	}

	// Execute the range query with the provided parameters
	ret, err := s.k.QueryPrometheusRange(query, startTime, endTime, step, timeout)
	if err != nil {
		errMsg := err.Error()
		// Check for common error patterns and provide more helpful messages
		if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
			return NewTextResult("", fmt.Errorf("ERROR: Metric not found. The specified metric '%s' does not exist in Prometheus. Please check the metric name and ensure it's correctly spelled.", query)), nil
		} else if strings.Contains(errMsg, "parse error") {
			return NewTextResult("", fmt.Errorf("ERROR: Invalid PromQL query syntax in '%s'. Please check your query format.", query)), nil
		} else if strings.Contains(errMsg, "failed to discover Prometheus") {
			return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
		} else if strings.Contains(errMsg, "invalid step") {
			return NewTextResult("", fmt.Errorf("ERROR: Invalid step parameter '%s'. Step must be a valid duration (e.g., '15s', '1m', '1h').", step)), nil
		} else if strings.Contains(errMsg, "resolution") || strings.Contains(errMsg, "step") {
			return NewTextResult("", fmt.Errorf("ERROR: Step parameter issue: %v. Adjust the step size or time range.", err)), nil
		}
		// For other errors, return a clear error message
		return NewTextResult("", fmt.Errorf("ERROR: Failed to execute Prometheus range query: %v", err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	return NewTextResult(ret, nil), nil
}

// prometheusListMetrics handles the prometheus_list_metrics tool request
func (s *Server) prometheusListMetrics(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ret, err := s.k.ListPrometheusMetrics()
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to list Prometheus metrics: %v", err)), nil
	}
	return NewTextResult(ret, nil), nil
}

// prometheusMetricInfo handles the prometheus_metric_info tool request
func (s *Server) prometheusMetricInfo(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required metric parameter
	metric := ctr.GetString("metric", "")
	if metric == "" {
		return NewTextResult("", errors.New("missing required parameter: metric")), nil
	}

	// Check if statistics are requested
	includeStats := false
	if statsArg := ctr.GetString("include_statistics", ""); statsArg != "" {
		includeStats = statsArg == "true"
	}

	ret, err := s.k.GetPrometheusMetricInfo(metric, includeStats)
	if err != nil {
		return NewTextResult("", fmt.Errorf("ERROR: Failed to get information for metric '%s': %v", metric, err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	return NewTextResult(ret, nil), nil
}

// prometheusGenerateQuery handles the prometheus_generate_query tool request
func (s *Server) prometheusGenerateQuery(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract the description parameter
	description := ctr.GetString("description", "")
	if description == "" {
		return NewTextResult("", errors.New("missing required parameter: description")), nil
	}

	// Generate the PromQL query
	query, err := s.k.GeneratePromQLQuery(description)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "failed to create LLM client") {
			return NewTextResult("", fmt.Errorf("ERROR: Could not connect to LLM service to generate PromQL query. The service may be unavailable.")), nil
		} else if strings.Contains(errMsg, "context deadline exceeded") || strings.Contains(errMsg, "timeout") {
			return NewTextResult("", fmt.Errorf("ERROR: Timeout occurred while generating the PromQL query. Please try again with a simpler description or try later.")), nil
		}
		return NewTextResult("", fmt.Errorf("ERROR: Failed to generate PromQL query from description: %v", err)), nil
	}

	// Check if the response is empty or too short to be a valid query
	if len(query) < 5 {
		return NewTextResult("", fmt.Errorf("ERROR: The generated query is too short or empty. Please provide a more specific description of the metric you're looking for.")), nil
	}

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
	// Extract the match parameter (required) using new API
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	matchArg, ok := argsMap["match"]
	if !ok {
		return NewTextResult("", errors.New("missing required parameter: match")), nil
	}

	// Convert the match parameter to a string slice
	matchSlice, ok := matchArg.([]interface{})
	if !ok {
		return NewTextResult("", errors.New("match parameter must be a string array")), nil
	}

	// Convert the match slice to a string slice
	match := make([]string, len(matchSlice))
	for i, m := range matchSlice {
		match[i], ok = m.(string)
		if !ok {
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

	// Call the Kubernetes function
	ret, err := s.k.QueryPrometheusSeries(match, startTime, endTime, limit)
	if err != nil {
		errMsg := err.Error()
		// Check for common error patterns and provide more helpful messages
		if strings.Contains(errMsg, "unknown by name") || strings.Contains(errMsg, "metrics not found") {
			return NewTextResult("", fmt.Errorf("ERROR: No series found matching the provided selectors. The metrics may not exist in Prometheus or may have different labels than specified.")), nil
		} else if strings.Contains(errMsg, "parse error") || strings.Contains(errMsg, "bad_data") {
			return NewTextResult("", fmt.Errorf("ERROR: Invalid series selector syntax in one of the match patterns: %v. Please check your selector format.", match)), nil
		} else if strings.Contains(errMsg, "failed to discover Prometheus") {
			return NewTextResult("", fmt.Errorf("ERROR: Cannot connect to Prometheus server. The server may be unavailable or misconfigured.")), nil
		}
		// For other errors, return a clear error message
		return NewTextResult("", fmt.Errorf("ERROR: Failed to query Prometheus series: %v", err)), nil
	}

	// Check if the response contains an ERROR_TYPE that indicates a conclusive empty result
	if strings.Contains(ret, "ERROR_TYPE") && (strings.Contains(ret, "NO_DATA_POINTS") || strings.Contains(ret, "NO_MATCHING_SERIES") || strings.Contains(ret, "METRIC_NOT_FOUND")) {
		// This is to ensure the model treats this as a definitive answer
		ret = "IMPORTANT - CONCLUSIVE RESULT: " + ret
	}

	return NewTextResult(ret, nil), nil
}

// prometheusTargets handles the prometheus_targets tool request
func (s *Server) prometheusTargets(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract optional state parameter
	state := ""
	if stateArg := ctr.GetString("state", ""); stateArg != "" {
		state = stateArg
	}

	// Extract optional scrape_pool parameter
	scrapePool := ""
	if scrapePoolArg := ctr.GetString("scrape_pool", ""); scrapePoolArg != "" {
		scrapePool = scrapePoolArg
	}

	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusTargets(state, scrapePool)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus targets: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusTargetMetadata handles the prometheus_targets_metadata tool request
func (s *Server) prometheusTargetMetadata(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract optional match_target parameter
	matchTarget := ""
	if matchTargetArg := ctr.GetString("match_target", ""); matchTargetArg != "" {
		matchTarget = matchTargetArg
	}

	// Extract optional metric parameter
	metric := ""
	if metricArg := ctr.GetString("metric", ""); metricArg != "" {
		metric = metricArg
	}

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

	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusTargetMetadata(matchTarget, metric, limit)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus target metadata: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

// Handler for creating Prometheus alerts
func (s *Server) prometheusCreateAlert(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate required parameters
	alertName := ctr.GetString("alertname", "")
	if alertName == "" {
		return NewTextResult("", errors.New("missing required parameter: alertname")), nil
	}

	expression := ctr.GetString("expression", "")
	if expression == "" {
		return NewTextResult("", errors.New("missing required parameter: expression")), nil
	}

	appLabel := ctr.GetString("applabel", "")
	if appLabel == "" {
		return NewTextResult("", errors.New("missing required parameter: applabel")), nil
	}

	namespace := ctr.GetString("namespace", "")
	if namespace == "" {
		return NewTextResult("", errors.New("missing required parameter: namespace")), nil
	}

	// Extract optional parameters with defaults
	var interval string
	if intervalArg := ctr.GetString("interval", ""); intervalArg != "" {
		interval = intervalArg
	} else {
		interval = "1m" // Default to 1 minute
	}

	var forDuration string
	if forArg := ctr.GetString("for", ""); forArg != "" {
		forDuration = forArg
	} else {
		forDuration = "5m" // Default to 5 minutes
	}

	// Convert annotations from interface{} to map[string]string using new API
	var annotations map[string]string
	args := ctr.GetRawArguments()
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
		result, err := s.k.CreatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration, annotations, alertLabels)
		if err != nil {
			return NewTextResult("", fmt.Errorf("failed to create Prometheus alert: %v", err)), nil
		}

		return NewTextResult(result, nil), nil
	}

	return NewTextResult("", errors.New("failed to get arguments")), nil
}

// Handler for updating Prometheus alerts
func (s *Server) prometheusUpdateAlert(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate required parameters
	alertName := ctr.GetString("alertname", "")
	if alertName == "" {
		return NewTextResult("", errors.New("missing required parameter: alertname")), nil
	}

	appLabel := ctr.GetString("applabel", "")
	if appLabel == "" {
		return NewTextResult("", errors.New("missing required parameter: applabel")), nil
	}

	namespace := ctr.GetString("namespace", "")
	if namespace == "" {
		return NewTextResult("", errors.New("missing required parameter: namespace")), nil
	}

	// Extract optional parameters
	var expression string
	if expressionArg := ctr.GetString("expression", ""); expressionArg != "" {
		expression = expressionArg
	}

	var interval string
	if intervalArg := ctr.GetString("interval", ""); intervalArg != "" {
		interval = intervalArg
	}

	var forDuration string
	if forArg := ctr.GetString("for", ""); forArg != "" {
		forDuration = forArg
	}

	// Convert annotations and alertlabels using new API
	var annotations map[string]string
	var alertLabels map[string]string

	args := ctr.GetRawArguments()
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
	result, err := s.k.UpdatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration, annotations, alertLabels)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to update Prometheus alert: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}

// Handler for deleting Prometheus alerts
func (s *Server) prometheusDeleteAlert(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate required parameters
	appLabel := ctr.GetString("applabel", "")
	if appLabel == "" {
		return NewTextResult("", errors.New("missing required parameter: applabel")), nil
	}

	namespace := ctr.GetString("namespace", "")
	if namespace == "" {
		return NewTextResult("", errors.New("missing required parameter: namespace")), nil
	}

	// Extract optional alertname parameter
	var alertName string
	if alertNameArg := ctr.GetString("alertname", ""); alertNameArg != "" {
		alertName = alertNameArg
	}

	// Call the Kubernetes function
	result, err := s.k.DeletePrometheusAlert(appLabel, namespace, alertName)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to delete Prometheus alert: %v", err)), nil
	}

	return NewTextResult(result, nil), nil
}

// prometheusGetAlerts handles the prometheus_get_alerts tool request
func (s *Server) prometheusGetAlerts(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusAlerts()
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus alerts: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusGetRules handles the prometheus_get_rules tool request
func (s *Server) prometheusGetRules(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract optional parameters using new API
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
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

	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusRules(groupLimit, ruleNames, ruleGroups, files, excludeAlerts, matchLabels)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus rules: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusCleanTombstones handles the prometheus_clean_tombstones tool request
func (s *Server) prometheusCleanTombstones(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call the Kubernetes function
	ret, err := s.k.CleanPrometheusTombstones()

	// If there's an error, we'll provide a user-friendly message and also include the
	// exact JSON response we know we're getting from the Prometheus server
	if err != nil {
		return NewTextResult("", fmt.Errorf("Cannot clean Prometheus tombstones: admin APIs are disabled on the Prometheus server. This is a security configuration that prevents administrative operations.\n\nServer response: {\"status\":\"error\",\"errorType\":\"unavailable\",\"error\":\"admin APIs disabled\"}")), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusCreateSnapshot handles the prometheus_create_snapshot tool request
func (s *Server) prometheusCreateSnapshot(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract optional skip_head parameter using GetBool
	skipHead := ctr.GetBool("skip_head", false)

	// Call the Kubernetes function
	ret, err := s.k.CreatePrometheusSnapshot(skipHead)

	// If there's an error, we'll provide a user-friendly message and also include the
	// exact JSON response we know we're getting from the Prometheus server
	if err != nil {
		return NewTextResult("", fmt.Errorf("Cannot create Prometheus snapshot: admin APIs are disabled on the Prometheus server. This is a security configuration that prevents administrative operations.\n\nServer response: {\"status\":\"error\",\"errorType\":\"unavailable\",\"error\":\"admin APIs disabled\"}")), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusDeleteSeries handles the prometheus_delete_series tool request
func (s *Server) prometheusDeleteSeries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required match parameter using GetRawArguments
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	matchArg, exists := argsMap["match"]
	if !exists || matchArg == nil {
		return NewTextResult("", errors.New("missing required parameter: match")), nil
	}

	// Convert the match parameter to a string slice
	matchSlice, ok := matchArg.([]interface{})
	if !ok {
		return NewTextResult("", errors.New("match parameter must be a string array")), nil
	}

	// Convert the match slice to a string slice
	match := make([]string, len(matchSlice))
	for i, m := range matchSlice {
		match[i], ok = m.(string)
		if !ok {
			return NewTextResult("", errors.New("match parameter must contain only strings")), nil
		}
	}

	// Extract optional start parameter
	var startTime *time.Time
	if startArg := ctr.GetString("start", ""); startArg != "" {
		parsed := parseTime(startArg, time.Time{})
		if !parsed.IsZero() {
			startTime = &parsed
		}
	}

	// Extract optional end parameter
	var endTime *time.Time
	if endArg := ctr.GetString("end", ""); endArg != "" {
		parsed := parseTime(endArg, time.Time{})
		if !parsed.IsZero() {
			endTime = &parsed
		}
	}

	// Call the Kubernetes function
	ret, err := s.k.DeletePrometheusSeries(match, startTime, endTime)

	// If there's an error, we'll provide a user-friendly message and also include the
	// exact JSON response we know we're getting from the Prometheus server
	if err != nil {
		return NewTextResult("", fmt.Errorf("Cannot delete Prometheus series: admin APIs are disabled on the Prometheus server. This is a security configuration that prevents administrative operations.\n\nServer response: {\"status\":\"error\",\"errorType\":\"unavailable\",\"error\":\"admin APIs disabled\"}")), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusAlertManagers handles the prometheus_alert_manager tool request
func (s *Server) prometheusAlertManagers(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusAlertManagers()
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus alertmanagers: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusRuntimeInfo handles the prometheus_runtimeinfo tool request
func (s *Server) prometheusRuntimeInfo(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusRuntimeInfo()
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus runtime info: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusTSDBStatus handles the prometheus_TSDB_status tool request
func (s *Server) prometheusTSDBStatus(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract optional limit parameter using GetFloat
	limit := int(ctr.GetFloat("limit", 0)) // Default is no limit

	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusTSDBStatus(limit)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus TSDB status: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

// prometheusWALReplay handles the prometheus_WALReplay tool request
func (s *Server) prometheusWALReplay(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call the Kubernetes function
	ret, err := s.k.GetPrometheusWALReplayStatus()
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get Prometheus WAL replay status: %v", err)), nil
	}

	// Check if we have empty data values (common case)
	if strings.Contains(ret, `"current": 0`) && strings.Contains(ret, `"max": 0`) && strings.Contains(ret, `"min": 0`) {
		return NewTextResult("WAL REPLAY STATUS: No active Write-Ahead Log (WAL) replay operations in progress. The 'current', 'max', and 'min' values are all 0, which indicates normal operation with no ongoing WAL replay activity. so no need to check for WAL replay status again.\n\n"+ret, nil), nil
	}

	// Add context to the response if there are active operations
	return NewTextResult("WAL REPLAY STATUS: Write-Ahead Log (WAL) replay status shows the progress of WAL replay operations. 'current' represents the current replay position, 'max' is the highest position, and 'min' is the lowest position.\n\n"+ret, nil), nil
}
