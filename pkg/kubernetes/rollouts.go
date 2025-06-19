package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ResourceRollout performs rollout operations on Kubernetes resources
func (k *Kubernetes) ResourceRollout(ctx context.Context, namespace, resourceType, resourceName, action string, revision int) (string, error) {
	resourceType = strings.ToLower(resourceType)
	action = strings.ToLower(action)

	// Ensure namespace is set
	namespace = namespaceOrDefault(namespace)

	// Create API endpoint based on action
	endpoint := fmt.Sprintf("/apis/v1/rollout-%s", action)

	// Add query parameters
	queryParams := url.Values{}
	queryParams.Add("namespace", namespace)
	queryParams.Add("resource_type", resourceType)
	queryParams.Add("resource_name", resourceName)

	if revision > 0 && action == "undo" {
		queryParams.Add("revision", strconv.Itoa(revision))
	}

	// Append query parameters to endpoint
	endpoint = endpoint + "?" + queryParams.Encode()

	// Make API request
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to %s rollout: %v", action, err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse rollout response: %v", err)
	}

	// Get formatted output if available, otherwise use message
	if formattedOutput, ok := result["formattedOutput"].(string); ok {
		return formattedOutput, nil
	} else if message, ok := result["message"].(string); ok {
		return message, nil
	} else {
		// Format JSON as string if no specific output format is available
		output, err := marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to format rollout response: %v", err)
		}
		return output, nil
	}
}
