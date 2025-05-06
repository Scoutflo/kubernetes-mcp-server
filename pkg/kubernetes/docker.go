package kubernetes

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/scoutflo/kubernetes-mcp-server/pkg/llm"
)

// DockerComposeToK8sManifest converts a Docker Compose file to a Kubernetes manifest
func (k *Kubernetes) DockerComposeToK8sManifest(dockerCompose string, namespace string) (string, error) {
	// Create a new LLM client
	llmClient, err := llm.NewDefaultClient()
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %v", err)
	}

	// Prepare user message with Docker Compose content
	userMessage := fmt.Sprintf("Please convert the following Docker Compose file to a single Kubernetes manifest file that contains all necessary resources. The manifest should include all Kubernetes objects needed to run the application defined in the Docker Compose file.\n\n```yaml\n%s\n```", dockerCompose)

	// Add namespace information if provided
	if namespace != "" {
		userMessage += fmt.Sprintf("\n\nPlease use '%s' as the namespace for all resources.", namespace)
	}

	// Make the LLM API call with the Docker Compose to K8s prompt and the Docker Compose content
	response, err := llmClient.Call(llm.DockerComposeToK8sManifestPrompt, userMessage)
	if err != nil {
		return "", fmt.Errorf("failed to convert Docker Compose to Kubernetes manifest: %v", err)
	}

	// Extract the YAML content from the response
	manifest := extractYAMLFromResponse(response)
	if manifest == "" {
		// If extraction failed, return the raw response
		return response, nil
	}

	return manifest, nil
}

// K8sManifestToHelmChart converts a Kubernetes manifest to a Helm chart
func (k *Kubernetes) K8sManifestToHelmChart(k8sManifest string, chartName string) (string, error) {
	// Create a new LLM client
	llmClient, err := llm.NewDefaultClient()
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %v", err)
	}

	// Prepare user message with Kubernetes manifest content
	userMessage := fmt.Sprintf("Please convert the following Kubernetes manifest into a complete Helm chart with all necessary files and templating. The chart should follow best practices and include all required components.\n\n```yaml\n%s\n```", k8sManifest)

	// Add chart name information if provided
	if chartName != "" {
		userMessage += fmt.Sprintf("\n\nPlease use '%s' as the name for the generated Helm chart.", chartName)
	}

	// Make the LLM API call with the K8s Manifest to Helm Chart prompt and the Kubernetes manifest content
	response, err := llmClient.Call(llm.K8sManifestToHelmChartPrompt, userMessage)
	if err != nil {
		return "", fmt.Errorf("failed to convert Kubernetes manifest to Helm chart: %v", err)
	}

	// Return the full response with structured Helm chart files
	return response, nil
}

// extractYAMLFromResponse extracts the YAML content from the LLM response
// It handles cases where the LLM might wrap the YAML in markdown code blocks
func extractYAMLFromResponse(response string) string {
	// Define regex patterns to match YAML code blocks
	yamlBlockPattern := regexp.MustCompile("(?s)```(?:yaml|yml)?\n(.*?)```")

	// Try to find YAML block with language tag
	matches := yamlBlockPattern.FindStringSubmatch(response)
	if len(matches) > 1 {
		return matches[1]
	}

	// If no match found with yaml tag, try finding any code block
	genericBlockPattern := regexp.MustCompile("(?s)```\n(.*?)```")
	matches = genericBlockPattern.FindStringSubmatch(response)
	if len(matches) > 1 {
		return matches[1]
	}

	// If still no match, check if the response looks like YAML
	if strings.Contains(response, "apiVersion:") && strings.Contains(response, "kind:") {
		return response
	}

	// If all else fails, return the original response
	return response
}
