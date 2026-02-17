package kubernetes

import (
	"context"
	"fmt"
)

// KialiHealthCheck checks the health of Kiali service
func (k *Kubernetes) KialiHealthCheck(ctx context.Context) (string, error) {
	response, err := k.MakeAPIRequest("GET", "/apis/v1/kiali/health", nil)
	if err != nil {
		return "", fmt.Errorf("failed to check Kiali health: %w", err)
	}
	return string(response), nil
}

// KialiStatus returns the status of Kiali installation
func (k *Kubernetes) KialiStatus(ctx context.Context) (string, error) {
	response, err := k.MakeAPIRequest("GET", "/apis/v1/kiali/status", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Kiali status: %w", err)
	}
	return string(response), nil
}

// GetNamespaceValidations gets Istio validations for a specific namespace
func (k *Kubernetes) GetNamespaceValidations(ctx context.Context, namespace string) (string, error) {
	if namespace == "" {
		return "", fmt.Errorf("namespace parameter is required")
	}

	endpoint := fmt.Sprintf("/apis/v1/kiali/namespaces/%s/validations", namespace)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get namespace validations for %s: %w", namespace, err)
	}
	return string(response), nil
}

// GetNamespaceTLS gets TLS status for a specific namespace
func (k *Kubernetes) GetNamespaceTLS(ctx context.Context, namespace string) (string, error) {
	if namespace == "" {
		return "", fmt.Errorf("namespace parameter is required")
	}

	endpoint := fmt.Sprintf("/apis/v1/kiali/namespaces/%s/tls", namespace)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get namespace TLS status for %s: %w", namespace, err)
	}
	return string(response), nil
}

// GetClustersServices gets services across clusters with optional namespace filtering
func (k *Kubernetes) GetClustersServices(ctx context.Context, namespace string) (string, error) {
	// Start with slim parameter for MCP to reduce payload size
	endpoint := "/apis/v1/kiali/clusters/services?fields=slim"

	// Add namespace filter if provided
	if namespace != "" {
		endpoint = fmt.Sprintf("%s&namespace=%s", endpoint, namespace)
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get clusters services: %w", err)
	}
	return string(response), nil
}

// GetKialiIstioConfig gets Istio configuration objects
func (k *Kubernetes) GetKialiIstioConfig(ctx context.Context) (string, error) {
	// Use slim parameter for MCP to reduce payload size
	response, err := k.MakeAPIRequest("GET", "/apis/v1/kiali/istio/config?fields=slim", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Istio config: %w", err)
	}
	return string(response), nil
}

// GetKialiTracingInfo gets tracing configuration and status
func (k *Kubernetes) GetKialiTracingInfo(ctx context.Context) (string, error) {
	response, err := k.MakeAPIRequest("GET", "/apis/v1/kiali/tracing", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get tracing info: %w", err)
	}
	return string(response), nil
}

// GetKialiNamespaceMetrics gets metrics for a namespace
func (k *Kubernetes) GetKialiNamespaceMetrics(ctx context.Context, namespace string) (string, error) {
	response, err := k.MakeAPIRequest("GET", fmt.Sprintf("/apis/v1/kiali/namespaces/%s/metrics", namespace), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get namespace metrics: %w", err)
	}
	return string(response), nil
}

// GetKialiServiceMetrics gets metrics for a service
func (k *Kubernetes) GetKialiServiceMetrics(ctx context.Context, namespace, service string) (string, error) {
	response, err := k.MakeAPIRequest("GET", fmt.Sprintf("/apis/v1/kiali/namespaces/%s/services/%s/metrics", namespace, service), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get service metrics: %w", err)
	}
	return string(response), nil
}

// GetKialiMeshGraph gets the mesh graph
func (k *Kubernetes) GetKialiMeshGraph(ctx context.Context) (string, error) {
	response, err := k.MakeAPIRequest("GET", "/apis/v1/kiali/mesh/graph", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get mesh graph: %w", err)
	}
	return string(response), nil
}
