package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) PodsListInAllNamespaces(ctx context.Context) (string, error) {
	return k.ResourcesList(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Pod",
	}, "")
}

func (k *Kubernetes) PodsListInNamespace(ctx context.Context, namespace string) (string, error) {
	return k.ResourcesList(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Pod",
	}, namespace)
}

func (k *Kubernetes) PodsGet(ctx context.Context, namespace, name string) (string, error) {
	return k.ResourcesGet(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Pod",
	}, namespaceOrDefault(namespace), name)
}

func (k *Kubernetes) PodsDelete(ctx context.Context, namespace, name string) (string, error) {
	queryParams := url.Values{}
	queryParams.Add("namespace", namespace)
	queryParams.Add("pod_name", name)

	endpoint := "/apis/v1/pod-delete?" + queryParams.Encode()

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to delete pod: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if message, ok := result["message"].(string); ok {
		return message, nil
	}

	return "Pod deleted successfully", nil
}

func (k *Kubernetes) PodsLog(ctx context.Context, namespace, name string, tailLines int) (string, error) {
	queryParams := url.Values{}
	queryParams.Add("namespace", namespace)
	queryParams.Add("pod_name", name)
	// Use the provided tailLines parameter
	queryParams.Add("tail_lines", fmt.Sprintf("%d", tailLines))

	endpoint := "/apis/v1/pod-logs?" + queryParams.Encode()

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if logs, ok := result["logs"].(string); ok {
		return logs, nil
	}

	return "", fmt.Errorf("logs not found in response")
}

func (k *Kubernetes) PodsRun(ctx context.Context, namespace, name, image string, port int32) (string, error) {
	// Create request body
	requestBody := map[string]interface{}{
		"namespace": namespaceOrDefault(namespace),
		"image":     image,
	}

	// Add name if specified
	if name != "" {
		requestBody["pod_name"] = name
	}

	// Add port if specified
	if port > 0 {
		requestBody["port"] = port
	}

	// Make API request - pass the map directly, not as JSON bytes
	response, err := k.MakeAPIRequest("POST", "/apis/v1/pod-run", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to run pod: %v", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Check for error
	if errMsg, ok := result["error"].(string); ok {
		return "", fmt.Errorf("pod run error: %s", errMsg)
	}

	// Format the resources information into YAML format
	if resources, ok := result["resources"].([]interface{}); ok && len(resources) > 0 {
		yamlOutput := ""
		for _, resource := range resources {
			if resourceMap, ok := resource.(map[string]interface{}); ok {
				kind := resourceMap["kind"].(string)
				name := resourceMap["name"].(string)
				namespace := resourceMap["namespace"].(string)

				yamlOutput += fmt.Sprintf("---\napiVersion: v1\nkind: %s\nmetadata:\n  name: %s\n  namespace: %s\n",
					kind, name, namespace)

				if kind == "Pod" {
					image := resourceMap["image"].(string)
					yamlOutput += fmt.Sprintf("spec:\n  containers:\n  - name: %s\n    image: %s\n", name, image)

					if port > 0 {
						yamlOutput += fmt.Sprintf("    ports:\n    - containerPort: %d\n", port)
					}
				} else if kind == "Service" {
					if portVal, ok := resourceMap["port"].(float64); ok {
						yamlOutput += fmt.Sprintf("spec:\n  selector:\n    app.kubernetes.io/name: %s\n  ports:\n  - port: %d\n    targetPort: %d\n",
							name, int(portVal), int(portVal))
					}
				}
			}
		}
		return yamlOutput, nil
	}

	return fmt.Sprintf("Pod %s created successfully in namespace %s",
		result["pod"].(string), result["namespace"].(string)), nil

}

func (k *Kubernetes) PodsExec(ctx context.Context, namespace, name, container string, command []string) (string, error) {
	// Create request body
	requestBody := map[string]interface{}{
		"namespace": namespaceOrDefault(namespace),
		"pod_name":  name,
		"command":   command,
	}

	// Add container if specified
	if container != "" {
		requestBody["container"] = container
	}

	// Make API request - pass the map directly, not as JSON bytes
	response, err := k.MakeAPIRequest("POST", "/apis/v1/pod-exec", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to execute command in pod: %v", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Check for error
	if errMsg, ok := result["error"].(string); ok {
		return "", fmt.Errorf("pod exec error: %s", errMsg)
	}

	// Return stdout if available
	if stdout, ok := result["stdout"].(string); ok {
		return stdout, nil
	}

	// Return stderr if stdout is empty
	if stderr, ok := result["stderr"].(string); ok {
		return stderr, nil
	}

	return "", fmt.Errorf("no output from command execution")
}
