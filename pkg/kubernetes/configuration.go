package kubernetes

import (
	"context"
	"fmt"
)

// GetAvailableAPIResources fetches all available API resources in the cluster
func GetAvailableAPIResources(ctx context.Context) (string, error) {
	// Create a new Kubernetes client
	k, err := NewKubernetes()
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Use the HTTP API to get API resources
	response, err := k.MakeAPIRequest("GET", "/apis/v1/api-resources", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get API resources: %v", err)
	}

	// The response is already in JSON format
	return string(response), nil
}
