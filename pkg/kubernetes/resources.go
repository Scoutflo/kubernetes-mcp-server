package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
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
	// Check if this is a Custom Resource
	if isCustomResource(gvk) {
		// For Custom Resources, we need to use a different approach
		// Since we can't directly list CRDs through our CRD functions, we'll use the API approach
		// but with proper CRD handling
		resourceType := getResourceTypeFromGVK(gvk)

		// Create request body for listing CRD resources
		requestBody := map[string]interface{}{
			"resource":  resourceType,
			"namespace": namespace,
			"action":    "list",
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body: %v", err)
		}

		// Use a generic API call for listing CRD resources
		response, err := k.MakeAPIRequest("POST", "/apis/v1/list-crd-resources", jsonData)
		if err != nil {
			return "", fmt.Errorf("failed to list custom resources: %v", err)
		}

		return string(response), nil
	}

	// For standard Kubernetes resources, use the existing logic
	resourceName := strings.ToLower(gvk.Kind)
	if gvk.Group != "" {
		resourceName = resourceName + "." + gvk.Group
	}

	// Build API endpoint
	endpoint := fmt.Sprintf("/api/v1/%s", resourceName)

	// Add namespace as query parameter if provided
	if namespace != "" {
		endpoint = endpoint + "?namespace=" + namespace
	}

	// Make API request
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	return string(response), nil
}

func (k *Kubernetes) ResourcesGet(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string) (string, error) {
	// Check if this is a Custom Resource
	if isCustomResource(gvk) {
		// Use CRD-specific function for Custom Resources
		resourceType := getResourceTypeFromGVK(gvk)

		// Use the GetCrdResource function
		result, err := k.GetCrdResource(resourceType, name, namespace)
		if err != nil {
			return "", fmt.Errorf("failed to get custom resource: %v", err)
		}

		// Convert result to JSON string
		jsonData, err := json.Marshal(result.Object)
		if err != nil {
			return "", fmt.Errorf("failed to marshal custom resource: %v", err)
		}

		return string(jsonData), nil
	}

	// For standard Kubernetes resources, use the existing logic
	resourceName := strings.ToLower(gvk.Kind)
	if gvk.Group != "" {
		resourceName = resourceName + "." + gvk.Group
	}

	// If namespace is empty, use default namespace
	namespace = namespaceOrDefault(namespace)

	// Build API endpoint
	endpoint := fmt.Sprintf("/api/v1/%s/%s/%s", resourceName, namespace, name)

	// Make API request
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get resource: %w", err)
	}

	return string(response), nil
}

func (k *Kubernetes) ResourcesCreateOrUpdate(ctx context.Context, resource string) (string, error) {
	// Parse the resource to determine if it's a Custom Resource
	var resourceObj map[string]interface{}

	// Try JSON parsing first
	if err := json.Unmarshal([]byte(resource), &resourceObj); err != nil {
		// Try YAML parsing if JSON fails
		if err := yaml.Unmarshal([]byte(resource), &resourceObj); err != nil {
			return "", fmt.Errorf("failed to parse resource as JSON or YAML: %v", err)
		}
	}

	// Extract apiVersion and kind to determine if it's a Custom Resource
	apiVersion, ok := resourceObj["apiVersion"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid apiVersion in resource")
	}

	kind, ok := resourceObj["kind"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid kind in resource")
	}

	// Parse the apiVersion to get Group and Version
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return "", fmt.Errorf("invalid apiVersion: %v", err)
	}

	gvk := &schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}

	// Check if this is a Custom Resource
	if isCustomResource(gvk) {
		// Use CRD-specific function for Custom Resources
		resourceType := getResourceTypeFromGVK(gvk)

		// Extract namespace from metadata if present
		namespace := ""
		if metadata, ok := resourceObj["metadata"].(map[string]interface{}); ok {
			if ns, ok := metadata["namespace"].(string); ok {
				namespace = ns
			}
		}

		// Try to create first, if it fails with "already exists", try update
		err := k.CreateCrdResource(resourceType, resourceObj, namespace)
		if err != nil {
			// If creation fails, try update
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "conflict") {
				err = k.UpdateCrdResource(resourceType, resourceObj, namespace)
				if err != nil {
					return "", fmt.Errorf("failed to update custom resource: %v", err)
				}
				return "Custom resource updated successfully", nil
			}
			return "", fmt.Errorf("failed to create custom resource: %v", err)
		}

		return "Custom resource created successfully", nil
	}

	// For standard Kubernetes resources, use the existing API logic
	requestBody := map[string]interface{}{
		"resource": resource,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Make API request to the resource-create-or-update endpoint
	response, err := k.MakeAPIRequest("POST", "/apis/v1/resource-create-or-update", jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to create or update resource: %v", err)
	}

	return string(response), nil
}

func (k *Kubernetes) ResourcesDelete(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string) error {
	// Check if this is a Custom Resource
	if isCustomResource(gvk) {
		// Use CRD-specific function for Custom Resources
		resourceType := getResourceTypeFromGVK(gvk)

		err := k.DeleteCrdResource(resourceType, name, namespace)
		if err != nil {
			return fmt.Errorf("failed to delete custom resource: %v", err)
		}

		return nil
	}

	// For standard Kubernetes resources, use the existing API logic
	requestBody := map[string]interface{}{
		"apiVersion": gvk.GroupVersion().String(),
		"kind":       gvk.Kind,
		"name":       name,
		"namespace":  namespace,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Make API request to the delete endpoint
	_, err = k.MakeAPIRequest("POST", "/apis/v1/delete", jsonData)
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

		// Convert to JSON
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body: %v", err)
		}

		// Make API request to the resource-patch endpoint
		response, err := k.MakeAPIRequest("POST", "/apis/v1/resource-patch", jsonData)
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

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Make API request to the resource-patch endpoint
	response, err := k.MakeAPIRequest("POST", "/apis/v1/resource-patch", jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to patch resource: %v", err)
	}

	return string(response), nil
}
