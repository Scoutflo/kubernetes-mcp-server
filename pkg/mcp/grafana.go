package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initGrafana() []server.ServerTool {
	return []server.ServerTool{
		{Tool: WithMeta(
			mcp.NewTool("grafana_health_check",
				mcp.WithDescription("Verify Grafana instance connectivity and API responsiveness. Call first when Grafana tools return unexpected errors — distinguishes a Grafana service outage from a datasource or dashboard-level issue."),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaHealthCheck},
		{Tool: WithMeta(
			mcp.NewTool("grafana_get_dashboard_by_uid",
				mcp.WithDescription("Retrieve complete dashboard configuration including panel layout, queries, and variables. Prefer over grafana_search_dashboards when uid is already known. The uid is a short alphanumeric string (e.g., 'abc123xyz') — not the dashboard title. Also required before any grafana_update_dashboard call to obtain the current JSON model with version field."),
				mcp.WithString("uid", mcp.Description("The UID of the dashboard — short alphanumeric string (e.g., 'abc123xyz'), not the display title"), mcp.Required()),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaGetDashboardByUID},
		{Tool: WithMeta(
			mcp.NewTool("grafana_search_dashboards",
				mcp.WithDescription("Search for Grafana dashboards by keyword. Use only when the dashboard uid is unknown — prefer grafana_get_dashboard_by_uid when uid is already known. Returns uid, title, and folder for each match."),
				mcp.WithString("query", mcp.Description("The query to search for")),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaSearchDashboards},
		{Tool: WithMeta(
			mcp.NewTool("grafana_update_dashboard",
				mcp.WithDescription("Create or update a Grafana dashboard. Write operation. Always call grafana_get_dashboard_by_uid first to obtain the current dashboard JSON — the update requires the full model including the 'version' field for optimistic locking. Omitting version or sending a stale version will fail or overwrite concurrent changes."),

				mcp.WithObject("dashboard",
					mcp.Description(`Full Grafana dashboard JSON model. Must include "title" and "panels" fields. Include "version" field from grafana_get_dashboard_by_uid for optimistic locking on updates. For new dashboards set "id" to null.`),
					mcp.Required()),

				mcp.WithString("folderUid",
					mcp.Description("Folder UID (leave empty for root folder)")),

				mcp.WithString("message",
					mcp.Description("Commit message for version history")),

				mcp.WithNumber("userId",
					mcp.Description("User ID for audit trail")),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaUpdateDashboard},
		{Tool: WithMeta(
			mcp.NewTool("grafana_get_dashboard_panel_queries",
				mcp.WithDescription("Retrieve the data source queries for all panels in a specific dashboard. Use to diagnose broken panels — reveals which datasource uid each panel queries and the exact query expression. More targeted than grafana_get_dashboard_by_uid when only panel query details are needed."),
				mcp.WithString("uid", mcp.Description("The UID of the dashboard — short alphanumeric string, not the display title"), mcp.Required()),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaGetDashboardPanelQueries},
		{Tool: WithMeta(
			mcp.NewTool("grafana_list_datasources",
				mcp.WithDescription("List all configured Grafana datasources with type and connection status. Use to discover datasource uid and name before calling grafana_get_datasource_by_uid or grafana_get_datasource_by_name. Filter by type (e.g., 'prometheus', 'loki') to narrow results."),
				mcp.WithString("type", mcp.Description("The type of datasources to search for. For example, 'prometheus', 'loki', 'tempo', etc...")),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaListDatasources},
		{Tool: WithMeta(
			mcp.NewTool("grafana_get_datasource_by_uid",
				mcp.WithDescription("Retrieve detailed connection parameters (url, type, access mode) for a datasource by uid. Prefer over grafana_get_datasource_by_name when uid is known. Use to diagnose datasource connectivity failures reported by panel queries."),
				mcp.WithString("uid", mcp.Description("The uid of the datasource — obtain from grafana_list_datasources when unknown"), mcp.Required()),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaGetDatasourceByUID},
		{Tool: WithMeta(
			mcp.NewTool("grafana_get_datasource_by_name",
				mcp.WithDescription("Retrieve datasource configuration by its registered name. Use when uid is unavailable — if uid is known, prefer grafana_get_datasource_by_uid. Name must match exactly as registered in Grafana (case-sensitive); use grafana_list_datasources to discover the exact name."),
				mcp.WithString("name", mcp.Description("The name of the datasource"), mcp.Required()),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaGetDatasourceByName},
		{Tool: WithMeta(
			mcp.NewTool("grafana_list_alert_rules",
				mcp.WithDescription("List all configured Grafana alert rules with evaluation status and notification policies. Use to discover alert rule uid when unknown — prefer grafana_get_alert_rule_by_uid when uid is already known. Supports label_selectors for filtering by alert labels."),
				mcp.WithNumber("limit", mcp.Description("The maximum number of results to return. Default is 100.")),
				mcp.WithNumber("page", mcp.Description("The page number to return.")),
				mcp.WithArray("label_selectors", mcp.Description("Optionally, a list of matchers to filter alert rules by labels. Each selector should have 'name', 'type' ('=' or '!='), and 'value' fields."),
					func(schema map[string]interface{}) {
						schema["type"] = "array"
						schema["items"] = map[string]interface{}{
							"type": "string",
						}
					},
				),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaListAlertRules},
		{Tool: WithMeta(
			mcp.NewTool("grafana_get_alert_rule_by_uid",
				mcp.WithDescription("Retrieve detailed configuration for a specific Grafana alert rule — condition query, datasource, 'for' duration, annotations, and notification policy. Prefer over grafana_list_alert_rules when uid is known. Use to inspect why an alert fired or is not firing."),
				mcp.WithString("uid", mcp.Description("The uid of the alert rule — obtain from grafana_list_alert_rules when unknown"), mcp.Required()),
			),
			map[string]any{"provider": ProviderGrafana},
		), Handler: s.grafanaGetAlertRuleByUID},
	}
}

// grafanaHealthCheck handles the grafana_health_check tool request
func (s *Server) grafanaHealthCheck(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: grafana_health_check - got called by session id: %s", sessionID)

	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_health_check failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	result, err := k.HealthCheck(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_health_check failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("Grafana health check failed: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: grafana_health_check completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaGetDashboardByUID handles the grafana_get_dashboard_by_uid tool request
func (s *Server) grafanaGetDashboardByUID(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_get_dashboard_by_uid failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_dashboard_by_uid failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_dashboard_by_uid - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the dashboard
	result, err := k.GetDashboardByUID(ctx, uid)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_get_dashboard_by_uid failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_get_dashboard_by_uid completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaSearchDashboards handles the grafana_search_dashboards tool request
func (s *Server) grafanaSearchDashboards(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_search_dashboards failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract optional query parameter
	query := ctr.GetString("query", "")

	klog.V(1).Infof("Tool: grafana_search_dashboards - query: %s - got called by session id: %s", query, sessionID)

	// Call the Kubernetes client to search dashboards
	result, err := k.SearchDashboards(ctx, query)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_search_dashboards failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_search_dashboards completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaUpdateDashboard handles the grafana_update_dashboard tool request
func (s *Server) grafanaUpdateDashboard(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_update_dashboard failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract parameters using GetRawArguments
	rawArgs := ctr.GetRawArguments()
	klog.Infof("[grafana_update_dashboard] Raw arguments: %#v", rawArgs)

	args, ok := rawArgs.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: arguments could not be cast to map[string]interface{} by session id: %s. Raw: %#v", time.Since(start), sessionID, rawArgs)
		return NewTextResult("", errors.New("arguments could not be cast to map[string]interface{}")), nil
	}

	dashboardArg, exists := args["dashboard"]
	klog.Infof("[grafana_update_dashboard] dashboard argument: %#v", dashboardArg)
	if !exists {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: missing required parameter: dashboard by session id: %s. Args: %#v", time.Since(start), sessionID, args)
		return NewTextResult("", errors.New("missing required parameter: dashboard")), nil
	}

	dashboard, ok := dashboardArg.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: dashboard parameter must be a JSON object by session id: %s. dashboardArg: %#v", time.Since(start), sessionID, dashboardArg)
		return NewTextResult("", errors.New("dashboard parameter must be a JSON object")), nil
	}

	// Validate that dashboard is not empty
	if len(dashboard) == 0 {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: dashboard parameter cannot be empty by session id: %s. dashboardArg: %#v", time.Since(start), sessionID, dashboardArg)
		return NewTextResult("", errors.New("dashboard parameter cannot be empty - must contain valid dashboard configuration with at least 'title' and 'panels' fields")), nil
	}

	// Validate required dashboard fields
	if _, hasTitle := dashboard["title"]; !hasTitle {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: dashboard missing required 'title' field by session id: %s. dashboard: %#v", time.Since(start), sessionID, dashboard)
		return NewTextResult("", errors.New("dashboard parameter must contain a 'title' field")), nil
	}

	// Validate that panels field exists (can be empty array but must be present)
	if _, hasPanels := dashboard["panels"]; !hasPanels {
		klog.Warningf("Tool call: grafana_update_dashboard - dashboard missing 'panels' field, adding empty panels array by session id: %s. dashboard: %#v", sessionID, dashboard)
		dashboard["panels"] = []interface{}{}
	}

	// Extract optional parameters
	folderUID := ctr.GetString("folderUid", "")
	message := ctr.GetString("message", "")
	overwrite := ctr.GetBool("overwrite", true)

	// Extract userID as number
	var userID int64
	if userIDArg, exists := args["userId"]; exists && userIDArg != nil {
		if userIDFloat, ok := userIDArg.(float64); ok {
			userID = int64(userIDFloat)
		}
	}

	// Log the dashboard structure for debugging
	dashboardJSON, _ := json.Marshal(dashboard)
	klog.V(1).Infof("Tool: grafana_update_dashboard - folderUID: %s, message: %s, overwrite: %t, userID: %d, dashboard_fields: %d, dashboard_json: %s - got called by session id: %s",
		folderUID, message, overwrite, userID, len(dashboard), string(dashboardJSON), sessionID)

	// Call the Kubernetes client to update the dashboard
	result, err := k.UpdateDashboard(ctx, dashboard, folderUID, message, overwrite, userID)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_update_dashboard completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaGetDashboardPanelQueries handles the grafana_get_dashboard_panel_queries tool request
func (s *Server) grafanaGetDashboardPanelQueries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	startTimeStr := ctr.GetString("start_time", "")
	endTimeStr := ctr.GetString("end_time", "")
	timeWindowStr := ctr.GetString("time_window", "")

	var startTime, endTime *time.Time
	if timeWindowStr != "" {
		duration, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed after %v: invalid time_window format: %v", time.Since(start), err)
			return NewTextResult("", fmt.Errorf("invalid time_window format '%s': %v", timeWindowStr, err)), nil
		}
		now := time.Now()
		start := now.Add(-duration)
		startTime = &start
		endTime = &now
	} else if startTimeStr != "" || endTimeStr != "" {
		if startTimeStr == "" || endTimeStr == "" {
			klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed after %v: both start_time and end_time must be provided together", time.Since(start))
			return NewTextResult("", errors.New("both start_time and end_time must be provided together, or use time_window")), nil
		}
		startParsed := parseTime(startTimeStr, time.Time{})
		if startParsed.IsZero() {
			klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed after %v: invalid start_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid start_time format, use RFC3339 or Unix timestamp")), nil
		}
		endParsed := parseTime(endTimeStr, time.Time{})
		if endParsed.IsZero() {
			klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed after %v: invalid end_time format", time.Since(start))
			return NewTextResult("", errors.New("invalid end_time format, use RFC3339 or Unix timestamp")), nil
		}
		startTime = &startParsed
		endTime = &endParsed
	}

	klog.V(1).Infof("Tool: grafana_get_dashboard_panel_queries - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the dashboard panel queries
	result, err := k.GetDashboardPanelQueries(ctx, uid, startTime, endTime)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_get_dashboard_panel_queries completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaListDatasources handles the grafana_list_datasources tool request
func (s *Server) grafanaListDatasources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_list_datasources failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract optional type parameter
	dsType := ctr.GetString("type", "")

	klog.V(1).Infof("Tool: grafana_list_datasources - type: %s - got called by session id: %s", dsType, sessionID)

	// Call the Kubernetes client to list datasources
	result, err := k.ListDatasources(ctx, dsType)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_list_datasources failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_list_datasources completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaGetDatasourceByUID handles the grafana_get_datasource_by_uid tool request
func (s *Server) grafanaGetDatasourceByUID(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_get_datasource_by_uid failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_datasource_by_uid failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_datasource_by_uid - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the datasource by UID
	result, err := k.GetDatasourceByUID(ctx, uid)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_get_datasource_by_uid failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_get_datasource_by_uid completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaGetDatasourceByName handles the grafana_get_datasource_by_name tool request
func (s *Server) grafanaGetDatasourceByName(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_get_datasource_by_name failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract required name parameter
	name := ctr.GetString("name", "")
	if name == "" {
		klog.Errorf("Tool call: grafana_get_datasource_by_name failed after %v: missing required parameter: name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: name")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_datasource_by_name - name: %s - got called by session id: %s", name, sessionID)

	// Call the Kubernetes client to get the datasource by name
	result, err := k.GetDatasourceByName(ctx, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_get_datasource_by_name failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_get_datasource_by_name completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaListAlertRules handles the list_alert_rules tool request
func (s *Server) grafanaListAlertRules(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_list_alert_rules failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract optional parameters
	args := ctr.GetRawArguments().(map[string]interface{})

	limit := 100 // default
	if limitArg, exists := args["limit"]; exists && limitArg != nil {
		if limitFloat, ok := limitArg.(float64); ok {
			limit = int(limitFloat)
		}
	}

	page := 1 // default
	if pageArg, exists := args["page"]; exists && pageArg != nil {
		if pageFloat, ok := pageArg.(float64); ok {
			page = int(pageFloat)
		}
	}

	var labelSelectors []map[string]interface{}
	if selectorsArg, exists := args["label_selectors"]; exists && selectorsArg != nil {
		if selectorsArray, ok := selectorsArg.([]interface{}); ok {
			for _, selectorInterface := range selectorsArray {
				if selector, ok := selectorInterface.(map[string]interface{}); ok {
					labelSelectors = append(labelSelectors, selector)
				}
			}
		}
	}

	klog.V(1).Infof("Tool: grafana_list_alert_rules - limit: %d, page: %d, label_selectors_count: %d - got called by session id: %s", limit, page, len(labelSelectors), sessionID)

	// Call the Kubernetes client to list alert rules
	result, err := k.ListAlertRules(ctx, limit, page, labelSelectors)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_list_alert_rules failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_list_alert_rules completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}

// grafanaGetAlertRuleByUID handles the get_alert_rule_by_uid tool request
func (s *Server) grafanaGetAlertRuleByUID(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: grafana_get_alert_rule_by_uid failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_alert_rule_by_uid failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_alert_rule_by_uid - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the alert rule by UID
	result, err := k.GetAlertRuleByUID(ctx, uid)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_get_alert_rule_by_uid failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_get_alert_rule_by_uid completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}
