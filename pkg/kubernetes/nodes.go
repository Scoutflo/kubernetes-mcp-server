package kubernetes

import (
	"context"
	"fmt"
)

// NodesList returns a list of all nodes in the cluster
func (k *Kubernetes) NodesList(ctx context.Context) (string, error) {
	// Make API request to list nodes
	response, err := k.MakeAPIRequest("GET", "/api/v1/nodes", nil)
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}
	return string(response), nil
}

// NodesGet returns detailed information about a specific node
func (k *Kubernetes) NodesGet(ctx context.Context, name string) (string, error) {
	// Make API request to get specific node
	endpoint := fmt.Sprintf("/api/v1/nodes//%s", name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get node %s: %w", name, err)
	}
	return string(response), nil
}
