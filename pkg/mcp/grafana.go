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
			mcp.WithDescription("Verify the health and connectivity of the Grafana instance to ensure operational reliability. This tool confirms API responsiveness and login functionality for effective monitoring and alerting. Validates in real-time."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaHealthCheck},
		{Tool: mcp.NewTool("grafana_get_dashboard_by_uid",
			mcp.WithDescription("Retrieve the full configuration of a specific Grafana dashboard to inspect its structure and settings. This tool diagnoses visualization issues impacting monitoring or alerting workflows."),
			mcp.WithString("uid", mcp.Description("The UID of the dashboard"), mcp.Required()),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaGetDashboardByUID},
		{Tool: mcp.NewTool("grafana_search_dashboards",
			mcp.WithDescription("Search for Grafana dashboards to locate relevant visualizations for monitoring or incident analysis. This tool identifies dashboards to support triage of alerts or performance issues."),
			mcp.WithString("query", mcp.Description("The query to search for")),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaSearchDashboards},
		{Tool: mcp.NewTool("grafana_update_dashboard",
			mcp.WithDescription("Update or create a Grafana dashboard with a specified configuration to enhance monitoring capabilities. This tool refines visualizations to address gaps in metrics or alerting workflows."),
			mcp.WithObject("dashboard", mcp.Description("The full dashboard JSON object containing all dashboard configuration"), mcp.Required()),
			mcp.WithString("folderUid", mcp.Description("The UID of the dashboard's folder (optional)")),
			mcp.WithString("message", mcp.Description("Set a commit message for the version history (optional)")),
			mcp.WithBoolean("overwrite", mcp.Description("Overwrite the dashboard if it exists. Otherwise create one (optional, default: false)")),
			mcp.WithNumber("userId", mcp.Description("ID of the user making the change (optional)")),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaUpdateDashboard},
		{Tool: mcp.NewTool("grafana_get_dashboard_panel_queries",
			mcp.WithDescription("Retrieve query details and datasource information for panels in a Grafana dashboard to analyze data retrieval issues. This tool supports troubleshooting by identifying misconfigured queries affecting visualizations."),
			mcp.WithString("uid", mcp.Description("The UID of the dashboard"), mcp.Required()),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaGetDashboardPanelQueries},
		{Tool: mcp.NewTool("grafana_list_datasources",
			mcp.WithDescription("List all configured Grafana datasources to verify the monitoring setup. This tool ensures datasources are active and properly configured for reliable metric collection."),
			mcp.WithString("type", mcp.Description("The type of datasources to search for. For example, 'prometheus', 'loki', 'tempo', etc...")),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaListDatasources},
		{Tool: mcp.NewTool("grafana_get_datasource_by_uid",
			mcp.WithDescription("Retrieve the configuration of a specific Grafana datasource by its UID to inspect connectivity settings. This tool diagnoses issues like incorrect URLs impacting data availability during troubleshooting."),
			mcp.WithString("uid", mcp.Description("The uid of the datasource"), mcp.Required()),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaGetDatasourceByUID},
		{Tool: mcp.NewTool("grafana_get_datasource_by_name",
			mcp.WithDescription("Retrieve the configuration of a Grafana datasource by its name to verify setup when the UID is unknown. This tool supports troubleshooting datasource connectivity for monitoring reliability."),
			mcp.WithString("name", mcp.Description("The name of the datasource"), mcp.Required()),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaGetDatasourceByName},
		{Tool: mcp.NewTool("grafana_list_alert_rules",
			mcp.WithDescription("List all Grafana alert rules with their status and configuration to review notification settings. This tool supports real-time triage of active alerts for incident response."),
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
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.grafanaListAlertRules},
		{Tool: mcp.NewTool("grafana_get_alert_rule_by_uid",
			mcp.WithDescription("Retrieve the configuration and status of a specific Grafana alert rule to analyze its behavior. This tool supports troubleshooting by refining rules to optimize alerting accuracy."),
			mcp.WithString("uid", mcp.Description("The uid of the alert rule"), mcp.Required()),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
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

	klog.V(1).Infof("Tool: grafana_get_dashboard_panel_queries - uid: %s - got called by session id: %s", uid, sessionID)

	// Call the Kubernetes client to get the dashboard panel queries
	result, err := k.GetDashboardPanelQueries(ctx, uid)
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
