package kubernetes

import (
	"context"
	"fmt"
)

// GetNodeMetrics returns CPU and memory metrics for all nodes or a specific node
func (k *Kubernetes) GetNodeMetrics(ctx context.Context, nodeName string) (string, error) {
	// Build API endpoint
	endpoint := "/apis/v1/metrics.k8s.io/v1beta1/nodes"

	// Add node name as query parameter if provided
	if nodeName != "" {
		endpoint = endpoint + "?node=" + nodeName
	}

	// Make API request to get node metrics
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get node metrics: %w", err)
	}

	return string(response), nil
}

// GetPodMetrics returns CPU and memory metrics for pods in a namespace
func (k *Kubernetes) GetPodMetrics(ctx context.Context, namespace string, podName string) (string, error) {
	// Build API endpoint with namespace
	endpoint := fmt.Sprintf("/apis/v1/metrics.k8s.io/v1beta1/namespaces/%s/pods", namespace)

	// Add pod name as query parameter if provided
	if podName != "" {
		endpoint = endpoint + "?pod=" + podName
	}

	// Make API request to get pod metrics
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get pod metrics: %w", err)
	}

	return string(response), nil
}
