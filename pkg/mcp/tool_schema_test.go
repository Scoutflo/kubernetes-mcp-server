package mcp

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func findTool(t *testing.T, tools []server.ServerTool, name string) mcp.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Tool.Name == name {
			return tool.Tool
		}
	}
	t.Fatalf("tool %q not found", name)
	return mcp.Tool{}
}

func requirePropertyType(t *testing.T, tool mcp.Tool, propertyName, expectedType string) map[string]any {
	t.Helper()
	property, ok := tool.InputSchema.Properties[propertyName].(map[string]any)
	if !ok {
		t.Fatalf("%s.%s property missing or wrong type: %#v", tool.Name, propertyName, tool.InputSchema.Properties[propertyName])
	}
	if got := property["type"]; got != expectedType {
		t.Fatalf("%s.%s type = %#v, want %q", tool.Name, propertyName, got, expectedType)
	}
	return property
}

func requireProperty(t *testing.T, tool mcp.Tool, propertyName string) map[string]any {
	t.Helper()
	property, ok := tool.InputSchema.Properties[propertyName].(map[string]any)
	if !ok {
		t.Fatalf("%s.%s property missing or wrong type: %#v", tool.Name, propertyName, tool.InputSchema.Properties[propertyName])
	}
	return property
}

func TestGrafanaToolSchemasExposeSupportedInputs(t *testing.T) {
	tools := (&Server{}).initGrafana()

	updateDashboard := findTool(t, tools, "grafana_update_dashboard")
	requirePropertyType(t, updateDashboard, "overwrite", "boolean")

	panelQueries := findTool(t, tools, "grafana_get_dashboard_panel_queries")
	requirePropertyType(t, panelQueries, "time_window", "string")
	requirePropertyType(t, panelQueries, "start_time", "string")
	requirePropertyType(t, panelQueries, "end_time", "string")

	alertRules := findTool(t, tools, "grafana_list_alert_rules")
	labelSelectors := requirePropertyType(t, alertRules, "label_selectors", "array")
	items, ok := labelSelectors["items"].(map[string]any)
	if !ok {
		t.Fatalf("grafana_list_alert_rules.label_selectors items missing: %#v", labelSelectors)
	}
	if got := items["type"]; got != "object" {
		t.Fatalf("grafana_list_alert_rules.label_selectors item type = %#v, want object", got)
	}
}

func TestCoreKubernetesToolSchemasExposeSupportedInputs(t *testing.T) {
	eventsList := findTool(t, (&Server{}).initEvents(), "events_list")
	requirePropertyType(t, eventsList, "time_window", "string")
	requirePropertyType(t, eventsList, "start_time", "string")
	requirePropertyType(t, eventsList, "end_time", "string")

	podsExec := findTool(t, (&Server{}).initPods(), "pods_exec")
	requirePropertyType(t, podsExec, "container", "string")

	podsLog := findTool(t, (&Server{}).initPods(), "pods_log")
	requirePropertyType(t, podsLog, "container", "string")

	resourcesPatch := findTool(t, (&Server{}).initResources(), "resources_patch")
	patch := requireProperty(t, resourcesPatch, "patch")
	if _, hasType := patch["type"]; hasType {
		t.Fatalf("resources_patch.patch should allow object or array payloads, got fixed type: %#v", patch)
	}
}

func TestPrometheusToolSchemasMatchHandlers(t *testing.T) {
	tools := (&Server{}).initPrometheus()

	seriesQuery := findTool(t, tools, "prometheus_series_query")
	requirePropertyType(t, seriesQuery, "time_window", "string")

	labelNames := findTool(t, tools, "prometheus_list_label_names")
	requirePropertyType(t, labelNames, "start", "string")
	requirePropertyType(t, labelNames, "end", "string")
	if _, exists := labelNames.InputSchema.Properties["startRfc3339"]; exists {
		t.Fatal("prometheus_list_label_names should expose start, not startRfc3339")
	}

	labelValues := findTool(t, tools, "prometheus_list_label_values")
	requirePropertyType(t, labelValues, "start", "string")
	requirePropertyType(t, labelValues, "end", "string")
	if _, exists := labelValues.InputSchema.Properties["endRfc3339"]; exists {
		t.Fatal("prometheus_list_label_values should expose end, not endRfc3339")
	}

	alerts := findTool(t, tools, "prometheus_get_alerts")
	requirePropertyType(t, alerts, "time_window", "string")
	requirePropertyType(t, alerts, "start_time", "string")
	requirePropertyType(t, alerts, "end_time", "string")

	rules := findTool(t, tools, "prometheus_get_rules")
	requirePropertyType(t, rules, "time_window", "string")
	requirePropertyType(t, rules, "start_time", "string")
	requirePropertyType(t, rules, "end_time", "string")
}
