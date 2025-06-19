package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// LabelResource applies labels to a Kubernetes resource by calling the K8s Dashboard API
func (k *Kubernetes) LabelResource(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string, labels map[string]string) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"apiVersion": gvk.GroupVersion().String(),
		"kind":       gvk.Kind,
		"name":       name,
		"namespace":  namespace,
		"labels":     labels,
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "apis/v1/label-resource", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to label resource: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	return "Labels applied successfully", nil
}

// RemoveLabel removes a label from a Kubernetes resource by calling the K8s Dashboard API
func (k *Kubernetes) RemoveLabel(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name, labelKey string) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"apiVersion": gvk.GroupVersion().String(),
		"kind":       gvk.Kind,
		"name":       name,
		"namespace":  namespace,
		"labelKey":   labelKey,
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "apis/v1/remove-label-resource", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to remove label: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	return "Label removed successfully", nil
}

// AnnotateResource applies annotations to a Kubernetes resource by calling the K8s Dashboard API
func (k *Kubernetes) AnnotateResource(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string, annotations map[string]string) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"apiVersion":  gvk.GroupVersion().String(),
		"kind":        gvk.Kind,
		"name":        name,
		"namespace":   namespace,
		"annotations": annotations,
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "apis/v1/annotate-resource", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to annotate resource: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	return "Annotations applied successfully", nil
}

// RemoveAnnotation removes an annotation from a Kubernetes resource by calling the K8s Dashboard API
func (k *Kubernetes) RemoveAnnotation(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name, annotationKey string) (string, error) {
	// Prepare request payload for API call
	requestPayload := map[string]interface{}{
		"apiVersion":    gvk.GroupVersion().String(),
		"kind":          gvk.Kind,
		"name":          name,
		"namespace":     namespace,
		"annotationKey": annotationKey,
	}

	// Make API call to K8s Dashboard
	response, err := k.MakeAPIRequest("POST", "apis/v1/remove-annotation-resource", requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to remove annotation: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	return "Annotation removed successfully", nil
}
