package mcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initPrometheus() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("prometheus_generate_query",
			mcp.WithDescription("Generate a PromQL query from a natural language description"),
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
			mcp.WithDescription("List all available Prometheus metrics with their metadata"),
		), Handler: s.prometheusListMetrics},
		{Tool: mcp.NewTool("prometheus_metric_info",
			mcp.WithDescription("Get detailed information about a specific Prometheus metric"),
			mcp.WithString("metric", mcp.Description("Name of the metric to get information about"), mcp.Required()),
			mcp.WithBoolean("include_statistics", mcp.Description("Include count, min, max, and avg statistics for this metric. May be slower for metrics with many time series.")),
		), Handler: s.prometheusMetricInfo},
	}
}

// prometheusMetrics handles the prometheus_metrics_query tool request
func (s *Server) prometheusMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required query parameter
	queryArg := ctr.Params.Arguments["query"]
	if queryArg == nil {
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	query := queryArg.(string)

	// Extract optional time parameter
	var queryTime *time.Time
	if timeArg := ctr.Params.Arguments["time"]; timeArg != nil {
		parsedTime := parseTime(timeArg.(string), time.Time{})
		if !parsedTime.IsZero() {
			queryTime = &parsedTime
		}
	}

	// Extract optional timeout parameter
	timeout := ""
	if timeoutArg := ctr.Params.Arguments["timeout"]; timeoutArg != nil {
		timeout = timeoutArg.(string)
	}

	// Execute the instant query with the provided parameters
	ret, err := s.k.QueryPrometheus(query, queryTime, timeout)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to execute Prometheus instant query: %v", err)), nil
	}
	return NewTextResult(ret, nil), nil
}

// prometheusMetricsRange handles the prometheus_metrics_query_range tool request
func (s *Server) prometheusMetricsRange(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	queryArg := ctr.Params.Arguments["query"]
	startArg := ctr.Params.Arguments["start"]
	endArg := ctr.Params.Arguments["end"]
	stepArg := ctr.Params.Arguments["step"]

	// Validate required parameters
	if queryArg == nil {
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	if startArg == nil {
		return NewTextResult("", errors.New("missing required parameter: start")), nil
	}
	if endArg == nil {
		return NewTextResult("", errors.New("missing required parameter: end")), nil
	}
	if stepArg == nil {
		return NewTextResult("", errors.New("missing required parameter: step")), nil
	}

	// Parse query
	query := queryArg.(string)

	// Parse start time
	startTime := parseTime(startArg.(string), time.Now().Add(-1*time.Hour))
	if startTime.IsZero() {
		return NewTextResult("", errors.New("invalid start time format")), nil
	}

	// Parse end time
	endTime := parseTime(endArg.(string), time.Now())
	if endTime.IsZero() {
		return NewTextResult("", errors.New("invalid end time format")), nil
	}

	// Parse step
	step := stepArg.(string)

	// Extract optional timeout parameter
	timeout := ""
	if timeoutArg := ctr.Params.Arguments["timeout"]; timeoutArg != nil {
		timeout = timeoutArg.(string)
	}

	// Execute the range query with the provided parameters
	ret, err := s.k.QueryPrometheusRange(query, startTime, endTime, step, timeout)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to execute Prometheus range query: %v", err)), nil
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
	metricArg := ctr.Params.Arguments["metric"]
	if metricArg == nil {
		return NewTextResult("", errors.New("missing required parameter: metric")), nil
	}
	metric := metricArg.(string)

	// Check if statistics are requested
	includeStats := false
	if statsArg := ctr.Params.Arguments["include_statistics"]; statsArg != nil {
		if b, ok := statsArg.(bool); ok {
			includeStats = b
		}
	}

	ret, err := s.k.GetPrometheusMetricInfo(metric, includeStats)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to get information for metric '%s': %v", metric, err)), nil
	}
	return NewTextResult(ret, nil), nil
}

// prometheusGenerateQuery handles the prometheus_generate_query tool request
func (s *Server) prometheusGenerateQuery(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract the description parameter
	descriptionArg := ctr.Params.Arguments["description"]
	if descriptionArg == nil {
		return NewTextResult("", errors.New("missing required parameter: description")), nil
	}
	description := descriptionArg.(string)

	// Generate the PromQL query
	query, err := s.k.GeneratePromQLQuery(description)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to generate PromQL query: %v", err)), nil
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
