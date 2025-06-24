package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initGrafana() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("grafana_health_check",
			mcp.WithDescription("Test basic connectivity to Grafana for debugging purposes"),
		), Handler: s.grafanaHealthCheck},
		{Tool: mcp.NewTool("grafana_get_dashboard_by_uid",
			mcp.WithDescription("Retrieves the complete dashboard, including panels, variables, and settings, for a specific dashboard identified by its UID."),
			mcp.WithString("uid", mcp.Description("The UID of the dashboard"), mcp.Required()),
		), Handler: s.grafanaGetDashboardByUID},
		{Tool: mcp.NewTool("grafana_search_dashboards",
			mcp.WithDescription("Search for Grafana dashboards by a query string. Returns a list of matching dashboards with details like title, UID, folder, tags, and URL."),
			mcp.WithString("query", mcp.Description("The query to search for")),
		), Handler: s.grafanaSearchDashboards},
		{Tool: mcp.NewTool("grafana_update_dashboard",
			mcp.WithDescription("Create or update a dashboard in Grafana. This tool allows you to save an existing dashboard with modifications or create a new one."),
			mcp.WithObject("dashboard", mcp.Description("The full dashboard JSON object containing all dashboard configuration"), mcp.Required()),
			mcp.WithString("folderUid", mcp.Description("The UID of the dashboard's folder (optional)")),
			mcp.WithString("message", mcp.Description("Set a commit message for the version history (optional)")),
			mcp.WithBoolean("overwrite", mcp.Description("Overwrite the dashboard if it exists. Otherwise create one (optional, default: false)")),
			mcp.WithNumber("userId", mcp.Description("ID of the user making the change (optional)")),
		), Handler: s.grafanaUpdateDashboard},
		{Tool: mcp.NewTool("grafana_get_dashboard_panel_queries",
			mcp.WithDescription("Get the title, query string, and datasource information for each panel in a dashboard. The datasource is an object with fields `uid` (which may be a concrete UID or a template variable like \"$datasource\") and `type`. If the datasource UID is a template variable, it won't be usable directly for queries. Returns an array of objects, each representing a panel, with fields: title, query, and datasource (an object with uid and type)."),
			mcp.WithString("uid", mcp.Description("The UID of the dashboard"), mcp.Required()),
		), Handler: s.grafanaGetDashboardPanelQueries},
		{Tool: mcp.NewTool("grafana_list_datasources",
			mcp.WithDescription("List available Grafana datasources. Optionally filter by datasource type (e.g., 'prometheus', 'loki'). Returns a summary list including ID, UID, name, type, and default status."),
			mcp.WithString("type", mcp.Description("The type of datasources to search for. For example, 'prometheus', 'loki', 'tempo', etc...")),
		), Handler: s.grafanaListDatasources},
		{Tool: mcp.NewTool("grafana_get_datasource_by_uid",
			mcp.WithDescription("Retrieves detailed information about a specific datasource using its UID. Returns the full datasource model, including name, type, URL, access settings, JSON data, and secure JSON field status."),
			mcp.WithString("uid", mcp.Description("The uid of the datasource"), mcp.Required()),
		), Handler: s.grafanaGetDatasourceByUID},
		{Tool: mcp.NewTool("grafana_get_datasource_by_name",
			mcp.WithDescription("Retrieves detailed information about a specific datasource using its name. Returns the full datasource model, including UID, type, URL, access settings, JSON data, and secure JSON field status."),
			mcp.WithString("name", mcp.Description("The name of the datasource"), mcp.Required()),
		), Handler: s.grafanaGetDatasourceByName},
		{Tool: mcp.NewTool("grafana_list_alert_rules",
			mcp.WithDescription("Lists Grafana alert rules, returning a summary including UID, title, current state (e.g., 'pending', 'firing', 'inactive'), and labels. Supports filtering by labels using selectors and pagination. Example label selector: `[{'name': 'severity', 'type': '=', 'value': 'critical'}]`. Inactive state means the alert state is normal, not firing."),
			mcp.WithNumber("limit", mcp.Description("The maximum number of results to return. Default is 100.")),
			mcp.WithNumber("page", mcp.Description("The page number to return.")),
			mcp.WithArray("label_selectors", mcp.Description("Optionally, a list of matchers to filter alert rules by labels. Each selector should have 'name', 'type' ('=' or '!='), and 'value' fields.")),
		), Handler: s.grafanaListAlertRules},
		{Tool: mcp.NewTool("grafana_get_alert_rule_by_uid",
			mcp.WithDescription("Retrieves the full configuration and detailed status of a specific Grafana alert rule identified by its unique ID (UID). The response includes fields like title, condition, query data, folder UID, rule group, state settings (no data, error), evaluation interval, annotations, and labels."),
			mcp.WithString("uid", mcp.Description("The uid of the alert rule"), mcp.Required()),
		), Handler: s.grafanaGetAlertRuleByUID},
	}
}

// grafanaHealthCheck handles the grafana_health_check tool request
func (s *Server) grafanaHealthCheck(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: grafana_health_check - got called by session id: %s", sessionID)

	result, err := s.k.HealthCheck()
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
	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_dashboard_by_uid failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_dashboard_by_uid - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the dashboard
	result, err := s.k.GetDashboardByUID(uid)
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
	// Extract optional query parameter
	query := ctr.GetString("query", "")

	klog.V(1).Infof("Tool: grafana_search_dashboards - query: %s - got called by session id: %s", query, sessionID)

	// Call the Kubernetes client to search dashboards
	result, err := s.k.SearchDashboards(query)
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
	// Extract parameters using GetRawArguments
	args := ctr.GetRawArguments().(map[string]interface{})

	// Extract required dashboard parameter
	dashboardArg, exists := args["dashboard"]
	if !exists {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: missing required parameter: dashboard by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: dashboard")), nil
	}

	dashboard, ok := dashboardArg.(map[string]interface{})
	if !ok {
		klog.Errorf("Tool call: grafana_update_dashboard failed after %v: dashboard parameter must be a JSON object by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("dashboard parameter must be a JSON object")), nil
	}

	// Extract optional parameters
	folderUID := ctr.GetString("folderUid", "")
	message := ctr.GetString("message", "")
	overwrite := ctr.GetBool("overwrite", false)

	// Extract userID as number
	var userID int64
	if userIDArg, exists := args["userId"]; exists && userIDArg != nil {
		if userIDFloat, ok := userIDArg.(float64); ok {
			userID = int64(userIDFloat)
		}
	}

	klog.V(1).Infof("Tool: grafana_update_dashboard - folderUID: %s, message: %s, overwrite: %t, userID: %d, dashboard_fields: %d - got called by session id: %s",
		folderUID, message, overwrite, userID, len(dashboard), sessionID)

	// Call the Kubernetes client to update the dashboard
	result, err := s.k.UpdateDashboard(dashboard, folderUID, message, overwrite, userID)
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
	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_dashboard_panel_queries failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_dashboard_panel_queries - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the dashboard panel queries
	result, err := s.k.GetDashboardPanelQueries(uid)
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
	// Extract optional type parameter
	dsType := ctr.GetString("type", "")

	klog.V(1).Infof("Tool: grafana_list_datasources - type: %s - got called by session id: %s", dsType, sessionID)

	// Call the Kubernetes client to list datasources
	result, err := s.k.ListDatasources(dsType)
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
	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_datasource_by_uid failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_datasource_by_uid - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the datasource by UID
	result, err := s.k.GetDatasourceByUID(uid)
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
	// Extract required name parameter
	name := ctr.GetString("name", "")
	if name == "" {
		klog.Errorf("Tool call: grafana_get_datasource_by_name failed after %v: missing required parameter: name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: name")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_datasource_by_name - name: %s - got called by session id: %s", name, sessionID)

	// Call the Kubernetes client to get the datasource by name
	result, err := s.k.GetDatasourceByName(name)
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
	result, err := s.k.ListAlertRules(limit, page, labelSelectors)
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
	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		klog.Errorf("Tool call: grafana_get_alert_rule_by_uid failed after %v: missing required parameter: uid by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	klog.V(1).Infof("Tool: grafana_get_alert_rule_by_uid - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the alert rule by UID
	result, err := s.k.GetAlertRuleByUID(uid)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: grafana_get_alert_rule_by_uid failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", err), nil
	}

	klog.V(1).Infof("Tool call: grafana_get_alert_rule_by_uid completed successfully in %v, result_length: %d by session id: %s", duration, len(result), sessionID)
	return NewTextResult(result, nil), nil
}
