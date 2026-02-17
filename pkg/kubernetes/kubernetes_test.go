package kubernetes

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
