package kubernetes

import (
	"context"
	"net/url"
	"strings"
)

func (k *Kubernetes) EventsList(ctx context.Context, namespace string, fieldSelectors []string) (string, error) {
	// Create the API endpoint URL with query parameters
	endpoint := "/apis/v1/get-events"

	// Add query parameters
	queryParams := url.Values{}

	if namespace != "" {
		queryParams.Add("namespace", namespace)
	}

	// Extract field selectors for involved objects
	for _, selector := range fieldSelectors {
		parts := strings.Split(selector, "=")
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "involvedObject.name":
			queryParams.Add("involved_object_name", value)
		case "involvedObject.kind":
			queryParams.Add("involved_object_kind", value)
		case "involvedObject.apiVersion":
			queryParams.Add("involved_object_api_version", value)
		}
	}

	// Append query parameters to the endpoint if any
	if len(queryParams) > 0 {
		endpoint = endpoint + "?" + queryParams.Encode()
	}

	// Make the API request
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	return string(response), nil
}
