package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// LabelResource applies labels to a Kubernetes resource
func (k *Kubernetes) LabelResource(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string, labels map[string]string) (string, error) {
	// Get the current resource
	resource, err := k.getResource(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting resource: %v", err)
	}

	// Prepare the JSON patch for labels
	patches := []map[string]interface{}{}

	// Make sure metadata.labels exists
	if resource.GetLabels() == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels",
			"value": map[string]string{},
		})
	}

	// Add each label as a separate patch operation
	for key, value := range labels {
		escapedKey := escapeJSONPointer(key)
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + escapedKey,
			"value": value,
		})
	}

	// Apply the patch
	patchData, err := json.Marshal(patches)
	if err != nil {
		return "", fmt.Errorf("error marshaling patch: %v", err)
	}

	if err := k.patchResource(ctx, gvk, namespace, name, types.JSONPatchType, patchData); err != nil {
		return "", fmt.Errorf("error applying labels: %v", err)
	}

	// Fetch the updated resource
	updated, err := k.ResourcesGet(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting updated resource: %v", err)
	}

	return fmt.Sprintf("Labels applied successfully to %s %s/%s\n%s", gvk.Kind, namespace, name, updated), nil
}

// RemoveLabel removes a label from a Kubernetes resource
func (k *Kubernetes) RemoveLabel(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name, labelKey string) (string, error) {
	// Get the current resource
	resource, err := k.getResource(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting resource: %v", err)
	}

	// Check if the resource has labels
	if resource.GetLabels() == nil {
		return "", fmt.Errorf("resource does not have any labels")
	}

	// Check if the label exists
	if _, exists := resource.GetLabels()[labelKey]; !exists {
		return "", fmt.Errorf("label with key '%s' does not exist on the resource", labelKey)
	}

	// Prepare the JSON patch to remove the label
	escapedKey := escapeJSONPointer(labelKey)
	patches := []map[string]interface{}{
		{
			"op":   "remove",
			"path": "/metadata/labels/" + escapedKey,
		},
	}

	// Apply the patch
	patchData, err := json.Marshal(patches)
	if err != nil {
		return "", fmt.Errorf("error marshaling patch: %v", err)
	}

	if err := k.patchResource(ctx, gvk, namespace, name, types.JSONPatchType, patchData); err != nil {
		return "", fmt.Errorf("error removing label: %v", err)
	}

	// Fetch the updated resource
	updated, err := k.ResourcesGet(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting updated resource: %v", err)
	}

	return fmt.Sprintf("Label '%s' successfully removed from %s %s/%s\n%s", labelKey, gvk.Kind, namespace, name, updated), nil
}

// AnnotateResource applies annotations to a Kubernetes resource
func (k *Kubernetes) AnnotateResource(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string, annotations map[string]string) (string, error) {
	// Get the current resource
	resource, err := k.getResource(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting resource: %v", err)
	}

	// Prepare the JSON patch for annotations
	patches := []map[string]interface{}{}

	// Make sure metadata.annotations exists
	if resource.GetAnnotations() == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/annotations",
			"value": map[string]string{},
		})
	}

	// Add each annotation as a separate patch operation
	for key, value := range annotations {
		escapedKey := escapeJSONPointer(key)
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/annotations/" + escapedKey,
			"value": value,
		})
	}

	// Apply the patch
	patchData, err := json.Marshal(patches)
	if err != nil {
		return "", fmt.Errorf("error marshaling patch: %v", err)
	}

	if err := k.patchResource(ctx, gvk, namespace, name, types.JSONPatchType, patchData); err != nil {
		return "", fmt.Errorf("error applying annotations: %v", err)
	}

	// Fetch the updated resource
	updated, err := k.ResourcesGet(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting updated resource: %v", err)
	}

	return fmt.Sprintf("Annotations applied successfully to %s %s/%s\n%s", gvk.Kind, namespace, name, updated), nil
}

// RemoveAnnotation removes an annotation from a Kubernetes resource
func (k *Kubernetes) RemoveAnnotation(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name, annotationKey string) (string, error) {
	// Get the current resource
	resource, err := k.getResource(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting resource: %v", err)
	}

	// Check if the resource has annotations
	if resource.GetAnnotations() == nil {
		return "", fmt.Errorf("resource does not have any annotations")
	}

	// Check if the annotation exists
	if _, exists := resource.GetAnnotations()[annotationKey]; !exists {
		return "", fmt.Errorf("annotation with key '%s' does not exist on the resource", annotationKey)
	}

	// Prepare the JSON patch to remove the annotation
	escapedKey := escapeJSONPointer(annotationKey)
	patches := []map[string]interface{}{
		{
			"op":   "remove",
			"path": "/metadata/annotations/" + escapedKey,
		},
	}

	// Apply the patch
	patchData, err := json.Marshal(patches)
	if err != nil {
		return "", fmt.Errorf("error marshaling patch: %v", err)
	}

	if err := k.patchResource(ctx, gvk, namespace, name, types.JSONPatchType, patchData); err != nil {
		return "", fmt.Errorf("error removing annotation: %v", err)
	}

	// Fetch the updated resource
	updated, err := k.ResourcesGet(ctx, gvk, namespace, name)
	if err != nil {
		return "", fmt.Errorf("error getting updated resource: %v", err)
	}

	return fmt.Sprintf("Annotation '%s' successfully removed from %s %s/%s\n%s", annotationKey, gvk.Kind, namespace, name, updated), nil
}

// Helper function to get a resource
func (k *Kubernetes) getResource(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string) (metav1.Object, error) {
	gvr, err := k.resourceFor(gvk)
	if err != nil {
		return nil, err
	}

	// If it's a namespaced resource and namespace wasn't provided, try to use the default configured one
	if namespaced, nsErr := k.isNamespaced(gvk); nsErr == nil && namespaced {
		namespace = namespaceOrDefault(namespace)
	}

	return k.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// Helper function to patch a resource
func (k *Kubernetes) patchResource(ctx context.Context, gvk *schema.GroupVersionKind, namespace, name string, patchType types.PatchType, patchData []byte) error {
	gvr, err := k.resourceFor(gvk)
	if err != nil {
		return err
	}

	// If it's a namespaced resource and namespace wasn't provided, try to use the default configured one
	if namespaced, nsErr := k.isNamespaced(gvk); nsErr == nil && namespaced {
		namespace = namespaceOrDefault(namespace)
	}

	_, err = k.dynamicClient.Resource(*gvr).Namespace(namespace).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	return err
}

// Helper function to escape JSON pointer path segments according to RFC 6901
func escapeJSONPointer(segment string) string {
	// Replace ~ with ~0 and / with ~1
	segment = strings.ReplaceAll(segment, "~", "~0")
	segment = strings.ReplaceAll(segment, "/", "~1")
	return segment
}
