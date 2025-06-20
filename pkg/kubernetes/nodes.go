package kubernetes

import (
	"context"
	"fmt"
)

// NodesList returns a list of all nodes in the cluster
func (k *Kubernetes) NodesList(ctx context.Context) (string, error) {
	// Create a JSON payload for the list-resources endpoint
	requestBody := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Node",
	}

	// Make API request to the dedicated MCP endpoint
	response, err := k.MakeAPIRequest("POST", "/apis/v1/list-resources", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}
	return string(response), nil
}

// NodesGet returns detailed information about a specific node
func (k *Kubernetes) NodesGet(ctx context.Context, name string) (string, error) {
	// Create a JSON payload for the get-resources endpoint
	requestBody := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Node",
		"name":       name,
	}

	// Make API request to the dedicated MCP endpoint
	response, err := k.MakeAPIRequest("POST", "/apis/v1/get-resources", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to get node %s: %w", name, err)
	}
	return string(response), nil
}
