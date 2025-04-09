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
		{Tool: mcp.NewTool("prometheus_metrics_query",
			mcp.WithDescription("Query Prometheus metrics using PromQL"),
			mcp.WithString("query", mcp.Description("PromQL query to execute"), mcp.Required()),
			mcp.WithString("start", mcp.Description("Start time for range queries (RFC3339 format or Unix timestamp). Optional, defaults to 1 hour ago for range queries.")),
			mcp.WithString("end", mcp.Description("End time for range queries (RFC3339 format or Unix timestamp). Optional, defaults to current time.")),
			mcp.WithString("step", mcp.Description("Step interval for range queries (e.g., '15s', '1m', '1h'). Optional, defaults to '15s'.")),
		), Handler: s.prometheusMetrics},
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

// prometheusMetrics handles the prometheus_metrics tool request
func (s *Server) prometheusMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required query parameter
	queryArg := ctr.Params.Arguments["query"]
	if queryArg == nil {
		return NewTextResult("", errors.New("missing required parameter: query")), nil
	}
	query := queryArg.(string)

	// Check if this is a range query by examining if start/end/step are provided
	startArg := ctr.Params.Arguments["start"]
	endArg := ctr.Params.Arguments["end"]
	stepArg := ctr.Params.Arguments["step"]

	// If any range parameters are provided, perform a range query
	if startArg != nil || endArg != nil || stepArg != nil {
		// Parse start time (default to 1 hour ago)
		var startTime time.Time
		if startArg != nil {
			startTime = parseTime(startArg.(string), time.Now().Add(-1*time.Hour))
		} else {
			startTime = time.Now().Add(-1 * time.Hour)
		}

		// Parse end time (default to now)
		var endTime time.Time
		if endArg != nil {
			endTime = parseTime(endArg.(string), time.Now())
		} else {
			endTime = time.Now()
		}

		// Parse step (default to 15s)
		step := "15s"
		if stepArg != nil {
			step = stepArg.(string)
		}

		// Execute the range query
		ret, err := s.k.QueryPrometheusRange(query, startTime, endTime, step)
		if err != nil {
			return NewTextResult("", fmt.Errorf("failed to execute Prometheus range query: %v", err)), nil
		}
		return NewTextResult(ret, nil), nil
	}

	// If no range parameters, perform an instant query
	ret, err := s.k.QueryPrometheus(query)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to execute Prometheus instant query: %v", err)), nil
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
