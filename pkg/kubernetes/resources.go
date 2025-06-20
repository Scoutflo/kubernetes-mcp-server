package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	AppKubernetesComponent = "app.kubernetes.io/component"
	AppKubernetesManagedBy = "app.kubernetes.io/managed-by"
	AppKubernetesName      = "app.kubernetes.io/name"
	AppKubernetesPartOf    = "app.kubernetes.io/part-of"
)

// isCustomResource checks if the given GVK represents a Custom Resource
func isCustomResource(gvk *schema.GroupVersionKind) bool {
	// Standard Kubernetes API groups that are NOT custom resources
	standardGroups := map[string]bool{
		"":                             true, // core/v1
		"v1":                           true, // core/v1
		"apps":                         true,
		"extensions":                   true,
		"networking.k8s.io":            true,
		"policy":                       true,
		"rbac.authorization.k8s.io":    true,
		"storage.k8s.io":               true,
		"batch":                        true,
		"autoscaling":                  true,
		"apiextensions.k8s.io":         true,
		"metrics.k8s.io":               true,
		"coordination.k8s.io":          true,
		"scheduling.k8s.io":            true,
		"admissionregistration.k8s.io": true,
		"node.k8s.io":                  true,
		"certificates.k8s.io":          true,
		"discovery.k8s.io":             true,
		"events.k8s.io":                true,
		"flowcontrol.apiserver.k8s.io": true,
	}

	// If the group is not in standard groups, it's likely a custom resource
	return !standardGroups[gvk.Group]
}

// getResourceTypeFromGVK converts GVK to resource type string for CRD functions
func getResourceTypeFromGVK(gvk *schema.GroupVersionKind) string {
	if gvk.Group == "" {
		return strings.ToLower(gvk.Kind)
	}
	return strings.ToLower(gvk.Kind) + "." + gvk.Group
}

func (k *Kubernetes) ResourcesList(ctx context.Context, gvk *schema.GroupVersionKind, namespace string) (string, error) {
	// Create a JSON payload for the list-resources endpoint
	requestBody := map[string]interface{}{
		"apiVersion": gvk.GroupVersion().String(),
		"kind":       gvk.Kind,
	}

	// Add namespace if provided
	if namespace != "" {
		requestBody["namespace"] = namespace
	}

	// Make API request to the dedicated MCP endpoint
	response, err := k.MakeAPIRequest("POST", "/apis/v1/list-resources", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to list resources: %v", err)
	}

	return string(response), nil
}

func (k *Kubernetes) ResourcesGet(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string) (string, error) {
	// Create a JSON payload for the get-resources endpoint
	requestBody := map[string]interface{}{
		"apiVersion": gvk.GroupVersion().String(),
		"kind":       gvk.Kind,
		"name":       name,
	}

	// Add namespace if provided, otherwise use default
	if namespace != "" {
		requestBody["namespace"] = namespace
	} else {
		requestBody["namespace"] = "default"
	}

	// Make API request to the dedicated MCP endpoint
	response, err := k.MakeAPIRequest("POST", "/apis/v1/get-resources", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to get resource: %v", err)
	}

	return string(response), nil
}

func (k *Kubernetes) ResourcesCreateOrUpdate(ctx context.Context, resource string) (string, error) {
	// Use the /apply endpoint which handles all resource types including Custom Resources
	// This is more robust than trying to detect and handle different resource types separately

	// The /apply endpoint expects raw YAML/JSON content in the request body
	// Make API request to the apply endpoint with the raw content
	response, err := k.MakeAPIRequest("POST", "/apis/v1/apply", []byte(resource))
	if err != nil {
		return "", fmt.Errorf("failed to apply resource: %v", err)
	}

	return string(response), nil
}

func (k *Kubernetes) ResourcesDelete(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string) error {
	// Create a common request format for all resource types
	requestBody := map[string]interface{}{
		"apiVersion": gvk.GroupVersion().String(),
		"kind":       gvk.Kind,
		"name":       name,
		"namespace":  namespace,
	}

	// Make API request to the delete endpoint
	_, err := k.MakeAPIRequest("POST", "/apis/v1/delete", requestBody)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %v", err)
	}

	return nil
}

// ResourcesPatch patches a resource using the K8s Dashboard API
func (k *Kubernetes) ResourcesPatch(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name, patchType string, patchData []byte) (string, error) {
	// Check if this is a Custom Resource
	if isCustomResource(gvk) {
		// For Custom Resources, we need to handle patching differently
		// Since we don't have a direct CRD patch function, we'll use the generic patch API
		// but identify it as a custom resource

		// Parse patch data back to interface for API request
		var patch interface{}
		if err := json.Unmarshal(patchData, &patch); err != nil {
			return "", fmt.Errorf("failed to unmarshal patch data: %v", err)
		}

		// Create request body for the API with CR flag
		requestBody := map[string]interface{}{
			"apiVersion":         gvk.GroupVersion().String(),
			"kind":               gvk.Kind,
			"resource_name":      name,
			"namespace":          namespace,
			"patch":              patch,
			"patch_type":         patchType,
			"is_custom_resource": true,
		}

		// Make API request to the resource-patch endpoint - pass the map directly
		response, err := k.MakeAPIRequest("POST", "/apis/v1/resource-patch", requestBody)
		if err != nil {
			return "", fmt.Errorf("failed to patch custom resource: %v", err)
		}

		return string(response), nil
	}

	// For standard Kubernetes resources, use the existing logic
	// Parse patch data back to interface for API request
	var patch interface{}
	if err := json.Unmarshal(patchData, &patch); err != nil {
		return "", fmt.Errorf("failed to unmarshal patch data: %v", err)
	}

	// Create request body for the API
	requestBody := map[string]interface{}{
		"apiVersion":    gvk.GroupVersion().String(),
		"kind":          gvk.Kind,
		"resource_name": name,
		"namespace":     namespace,
		"patch":         patch,
		"patch_type":    patchType,
	}

	// Make API request to the resource-patch endpoint - pass the map directly
	response, err := k.MakeAPIRequest("POST", "/apis/v1/resource-patch", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to patch resource: %v", err)
	}

	return string(response), nil
}
