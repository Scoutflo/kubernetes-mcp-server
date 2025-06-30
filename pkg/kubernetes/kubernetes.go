package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// HTTPClient represents an HTTP client for communicating with K8s Dashboard API
type HTTPClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewHTTPClient creates a new HTTP client for K8s Dashboard API
func NewHTTPClient(baseURL, token string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Token:   token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MakeRequest makes an HTTP request to the K8s Dashboard API
func (h *HTTPClient) MakeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	var contentType string = "application/json"

	if body != nil {
		// Check if body is already raw bytes (for YAML content)
		if rawBytes, ok := body.([]byte); ok {
			reqBody = bytes.NewBuffer(rawBytes)
			contentType = "application/yaml"
		} else {
			// Marshal as JSON for structured data
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		}
	}

	url := strings.TrimSuffix(h.BaseURL, "/") + "/" + strings.TrimPrefix(endpoint, "/")
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+h.Token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

type Kubernetes struct {
	// HTTP Client for K8s Dashboard API
	HTTPClient *HTTPClient
}

// NewKubernetesWithCredentials creates a new Kubernetes client with the provided URL and token
func NewKubernetesWithCredentials(k8sURL, k8sToken string) (*Kubernetes, error) {
	if k8sURL == "" || k8sToken == "" {
		return nil, fmt.Errorf("k8sURL and k8sToken must be provided")
	}

	k8s := &Kubernetes{}

	// Initialize HTTP client mode
	k8s.HTTPClient = NewHTTPClient(k8sURL, k8sToken)

	// Test the connection with a health check
	_, err := k8s.HTTPClient.MakeRequest("GET", "/healthz", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to K8s Dashboard API: %w", err)
	}

	return k8s, nil
}

func marshal(v any) (string, error) {
	switch t := v.(type) {
	case []unstructured.Unstructured:
		for i := range t {
			t[i].SetManagedFields(nil)
		}
	case []*unstructured.Unstructured:
		for i := range t {
			t[i].SetManagedFields(nil)
		}
	case unstructured.Unstructured:
		t.SetManagedFields(nil)
	case *unstructured.Unstructured:
		t.SetManagedFields(nil)
	}
	ret, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// namespaceOrDefault returns the provided namespace or "default" if empty
func namespaceOrDefault(namespace string) string {
	if namespace == "" {
		return "default"
	}
	return namespace
}

// MakeAPIRequest is a convenience method to make API requests to K8s Dashboard API
func (k *Kubernetes) MakeAPIRequest(method, endpoint string, body interface{}) ([]byte, error) {
	return k.HTTPClient.MakeRequest(method, endpoint, body)
}
