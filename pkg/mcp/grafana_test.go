package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGrafanaListAlertRulesAllowsMissingOptionalArguments(t *testing.T) {
	var sawAlertRules bool
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/apis/v1/grafana/alert-rules":
			sawAlertRules = true
			if got := r.URL.Query().Get("fields"); got != "slim" {
				t.Fatalf("fields query = %q, want slim", got)
			}
			if got := r.URL.Query().Get("limit"); got != "100" {
				t.Fatalf("limit query = %q, want 100", got)
			}
			if got := r.URL.Query().Get("page"); got != "1" {
				t.Fatalf("page query = %q, want 1", got)
			}
			_ = json.NewEncoder(w).Encode([]map[string]string{{"uid": "alert-1"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	ctr := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "grafana_list_alert_rules",
			Meta: mcp.NewMetaFromMap(map[string]any{
				"k8surl":   backend.URL,
				"k8stoken": "test-token",
			}),
		},
	}

	result, err := (&Server{}).grafanaListAlertRules(context.Background(), ctr)
	if err != nil {
		t.Fatalf("grafanaListAlertRules returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("grafanaListAlertRules returned tool error: %#v", result.Content)
	}
	if !sawAlertRules {
		t.Fatal("expected alert-rules endpoint to be called")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected text content")
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want mcp.TextContent", result.Content[0])
	}
	if !strings.Contains(text.Text, "alert-1") {
		t.Fatalf("response text = %q, want alert uid", text.Text)
	}
}
