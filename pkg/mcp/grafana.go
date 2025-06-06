package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	}
}

// grafanaHealthCheck handles the grafana_health_check tool request
func (s *Server) grafanaHealthCheck(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := s.k.HealthCheck()
	if err != nil {
		return NewTextResult("", fmt.Errorf("Grafana health check failed: %v", err)), nil
	}
	return NewTextResult(result, nil), nil
}

// grafanaGetDashboardByUID handles the grafana_get_dashboard_by_uid tool request
func (s *Server) grafanaGetDashboardByUID(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	// Call the Kubernetes client to get the dashboard
	result, err := s.k.GetDashboardByUID(uid)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// grafanaSearchDashboards handles the grafana_search_dashboards tool request
func (s *Server) grafanaSearchDashboards(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract optional query parameter
	query := ctr.GetString("query", "")

	// Call the Kubernetes client to search dashboards
	result, err := s.k.SearchDashboards(query)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// grafanaUpdateDashboard handles the grafana_update_dashboard tool request
func (s *Server) grafanaUpdateDashboard(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters using GetRawArguments
	args := ctr.GetRawArguments().(map[string]interface{})

	// Extract required dashboard parameter
	dashboardArg, exists := args["dashboard"]
	if !exists {
		return NewTextResult("", errors.New("missing required parameter: dashboard")), nil
	}

	dashboard, ok := dashboardArg.(map[string]interface{})
	if !ok {
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

	// Call the Kubernetes client to update the dashboard
	result, err := s.k.UpdateDashboard(dashboard, folderUID, message, overwrite, userID)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// grafanaGetDashboardPanelQueries handles the grafana_get_dashboard_panel_queries tool request
func (s *Server) grafanaGetDashboardPanelQueries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	// Call the Kubernetes client to get the dashboard panel queries
	result, err := s.k.GetDashboardPanelQueries(uid)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// grafanaListDatasources handles the grafana_list_datasources tool request
func (s *Server) grafanaListDatasources(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract optional type parameter
	dsType := ctr.GetString("type", "")

	// Call the Kubernetes client to list datasources
	result, err := s.k.ListDatasources(dsType)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// grafanaGetDatasourceByUID handles the grafana_get_datasource_by_uid tool request
func (s *Server) grafanaGetDatasourceByUID(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required uid parameter
	uid := ctr.GetString("uid", "")
	if uid == "" {
		return NewTextResult("", errors.New("missing required parameter: uid")), nil
	}

	// Call the Kubernetes client to get the datasource by UID
	result, err := s.k.GetDatasourceByUID(uid)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}

// grafanaGetDatasourceByName handles the grafana_get_datasource_by_name tool request
func (s *Server) grafanaGetDatasourceByName(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required name parameter
	name := ctr.GetString("name", "")
	if name == "" {
		return NewTextResult("", errors.New("missing required parameter: name")), nil
	}

	// Call the Kubernetes client to get the datasource by name
	result, err := s.k.GetDatasourceByName(name)
	if err != nil {
		return NewTextResult("", err), nil
	}

	return NewTextResult(result, nil), nil
}
