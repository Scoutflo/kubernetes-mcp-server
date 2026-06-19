package kubernetes

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlimTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		initialQuery   string
		skipSlimHeader bool
		wantFields     string
	}{
		{
			name:       "GET request adds fields=slim",
			method:     http.MethodGet,
			wantFields: "slim",
		},
		{
			name:         "GET request preserves existing fields value",
			method:       http.MethodGet,
			initialQuery: "fields=full",
			wantFields:   "full",
		},
		{
			name:         "GET request preserves other query params",
			method:       http.MethodGet,
			initialQuery: "namespace=default",
			wantFields:   "slim",
		},
		{
			name:           "GET request with skip header does not add fields",
			method:         http.MethodGet,
			skipSlimHeader: true,
			wantFields:     "",
		},
		{
			name:       "POST request does not add fields",
			method:     http.MethodPost,
			wantFields: "",
		},
		{
			name:       "PUT request does not add fields",
			method:     http.MethodPut,
			wantFields: "",
		},
		{
			name:       "DELETE request does not add fields",
			method:     http.MethodDelete,
			wantFields: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedReq *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			transport := &SlimTransport{Base: http.DefaultTransport}

			url := server.URL
			if tt.initialQuery != "" {
				url += "?" + tt.initialQuery
			}

			req, err := http.NewRequest(tt.method, url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			if tt.skipSlimHeader {
				req.Header.Set(SkipSlimHeader, "true")
			}

			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("RoundTrip failed: %v", err)
			}
			defer resp.Body.Close()

			if capturedReq == nil {
				t.Fatal("request was not captured")
			}

			gotFields := capturedReq.URL.Query().Get("fields")
			if gotFields != tt.wantFields {
				t.Errorf("fields = %q, want %q", gotFields, tt.wantFields)
			}

			if tt.skipSlimHeader && capturedReq.Header.Get(SkipSlimHeader) != "" {
				t.Error("SkipSlimHeader was not removed from request")
			}
		})
	}
}

func TestSlimTransport_NilBase(t *testing.T) {
	transport := &SlimTransport{Base: nil}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fields") != "slim" {
			t.Errorf("fields = %q, want %q", r.URL.Query().Get("fields"), "slim")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestNewHTTPClient_UsesSlimTransport(t *testing.T) {
	client := NewHTTPClient("http://localhost:8080", "test-token")

	if _, ok := client.Client.Transport.(*SlimTransport); !ok {
		t.Error("HTTPClient does not use SlimTransport")
	}
}

func TestGetPrometheusRules_SendsStringTimes(t *testing.T) {
	var body map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apis/v1/prometheus/rules" {
			t.Fatalf("path = %q, want /apis/v1/prometheus/rules", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := &Kubernetes{HTTPClient: NewHTTPClient(server.URL, "test-token")}
	start := time.Date(2026, 6, 19, 12, 1, 24, 0, time.UTC)
	end := start.Add(15 * time.Minute)

	if _, err := client.GetPrometheusRules(0, []string{"AfterdoQAPodCrashLoop"}, nil, nil, false, nil, &start, &end); err != nil {
		t.Fatalf("GetPrometheusRules failed: %v", err)
	}

	if got, ok := body["start_time"].(string); !ok || got != "2026-06-19T12:01:24Z" {
		t.Fatalf("start_time = %#v, want RFC3339 string", body["start_time"])
	}
	if got, ok := body["end_time"].(string); !ok || got != "2026-06-19T12:16:24Z" {
		t.Fatalf("end_time = %#v, want RFC3339 string", body["end_time"])
	}
}
