package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// GetRollout gets an Argo Rollout by name and namespace using K8s Dashboard API
func (k *Kubernetes) GetRollout(ctx context.Context, name, namespace string) (*rolloutv1alpha1.Rollout, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Construct the API endpoint
	endpoint := fmt.Sprintf("/api/v1/argorollouts/rollouts/%s/%s", namespace, name)

	// Make API request using k.MakeAPIRequest
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if errMsg, ok := result["error"].(string); ok {
		return nil, fmt.Errorf("API error: %s", errMsg)
	}

	// Extract the rollout from the response
	rolloutData, ok := result["rollout"]
	if !ok {
		return nil, fmt.Errorf("rollout data not found in response")
	}

	// Convert back to rollout object
	rolloutJSON, err := json.Marshal(rolloutData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rollout data: %w", err)
	}

	var rollout rolloutv1alpha1.Rollout
	if err := json.Unmarshal(rolloutJSON, &rollout); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rollout: %w", err)
	}

	return &rollout, nil
}

// PromoteRollout promotes an Argo Rollout to advance it to the next step
func (k *Kubernetes) PromoteRollout(ctx context.Context, name, namespace string, fullPromote bool) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Construct the API endpoint for promoting rollout
	endpoint := fmt.Sprintf("/api/v1/argorollouts/rollouts/%s/%s/promote", namespace, name)

	// Prepare request body
	requestBody := map[string]interface{}{
		"full": fullPromote,
	}

	// Make API request
	response, err := k.MakeAPIRequest("POST", endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to promote rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if errMsg, ok := result["error"].(string); ok {
		return "", fmt.Errorf("API error: %s", errMsg)
	}

	// Extract the result message
	if message, ok := result["message"].(string); ok {
		log.Infof("Successfully promoted rollout '%s' in namespace '%s'", name, namespace)
		return message, nil
	}

	return fmt.Sprintf("Rollout '%s' in namespace '%s' has been promoted", name, namespace), nil
}

// AbortRollout aborts an in-progress Argo Rollout and reverts to the stable version
func (k *Kubernetes) AbortRollout(ctx context.Context, name, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Construct the API endpoint for aborting rollout
	endpoint := fmt.Sprintf("/api/v1/argorollouts/rollouts/%s/%s/abort", namespace, name)

	// Make API request
	response, err := k.MakeAPIRequest("POST", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to abort rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if errMsg, ok := result["error"].(string); ok {
		return "", fmt.Errorf("API error: %s", errMsg)
	}

	// Extract the result message
	if message, ok := result["message"].(string); ok {
		log.Infof("Successfully aborted rollout '%s' in namespace '%s'", name, namespace)
		return message, nil
	}

	return fmt.Sprintf("Rollout '%s' in namespace '%s' has been aborted", name, namespace), nil
}

// SetRolloutWeight sets the weight for a canary rollout
func (k *Kubernetes) SetRolloutWeight(ctx context.Context, name, namespace string, weight int) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	if weight < 0 || weight > 100 {
		return "", fmt.Errorf("weight must be between 0 and 100")
	}

	// Construct the API endpoint for setting rollout weight
	endpoint := fmt.Sprintf("/api/v1/argorollouts/rollouts/%s/%s/weight", namespace, name)

	// Prepare request body
	requestBody := map[string]interface{}{
		"weight": weight,
	}

	// Make API request
	response, err := k.MakeAPIRequest("POST", endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to set canary weight for rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if errMsg, ok := result["error"].(string); ok {
		return "", fmt.Errorf("API error: %s", errMsg)
	}

	// Extract the result message
	if message, ok := result["message"].(string); ok {
		log.Infof("Successfully set canary weight to %d%% for rollout '%s' in namespace '%s'", weight, name, namespace)
		return message, nil
	}

	return fmt.Sprintf("Canary weight for rollout '%s' in namespace '%s' has been set to %d%%", name, namespace, weight), nil
}

// FormatRolloutOutput formats rollout output based on the requested format
func (k *Kubernetes) FormatRolloutOutput(rollout *rolloutv1alpha1.Rollout, format string) (string, error) {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(rollout, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal rollout to JSON: %w", err)
		}
		return string(jsonData), nil
	case "yaml":
		yamlData, err := marshal(rollout)
		if err != nil {
			return "", fmt.Errorf("failed to marshal rollout to YAML: %w", err)
		}
		return yamlData, nil
	case "wide":
		return k.formatRolloutWide(rollout), nil
	default:
		return k.formatRolloutDefault(rollout), nil
	}
}

// Helper functions for formatting rollout information
func (k *Kubernetes) getStrategyType(rollout *rolloutv1alpha1.Rollout) string {
	if rollout.Spec.Strategy.BlueGreen != nil {
		return "BlueGreen"
	}
	if rollout.Spec.Strategy.Canary != nil {
		return "Canary"
	}
	return "Unknown"
}

func (k *Kubernetes) getContainerImages(rollout *rolloutv1alpha1.Rollout) string {
	var images []string
	for _, container := range rollout.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}
	return strings.Join(images, ", ")
}

func (k *Kubernetes) formatRolloutDefault(rollout *rolloutv1alpha1.Rollout) string {
	return fmt.Sprintf("NAME: %s\nNAMESPACE: %s\nSTRATEGY: %s\nIMAGES: %s\nSTATUS: %s\nREADY: %d/%d",
		rollout.Name,
		rollout.Namespace,
		k.getStrategyType(rollout),
		k.getContainerImages(rollout),
		rollout.Status.Phase,
		rollout.Status.ReadyReplicas,
		*rollout.Spec.Replicas,
	)
}

func (k *Kubernetes) formatRolloutWide(rollout *rolloutv1alpha1.Rollout) string {
	return fmt.Sprintf("NAME: %s\nNAMESPACE: %s\nSTRATEGY: %s\nIMAGES: %s\nSTATUS: %s\nREADY: %d/%d\nUPDATED: %d\nAVAILABLE: %d",
		rollout.Name,
		rollout.Namespace,
		k.getStrategyType(rollout),
		k.getContainerImages(rollout),
		rollout.Status.Phase,
		rollout.Status.ReadyReplicas,
		*rollout.Spec.Replicas,
		rollout.Status.UpdatedReplicas,
		rollout.Status.AvailableReplicas,
	)
}

// GenerateRolloutYAML generates a YAML configuration for an Argo Rollout
func (k *Kubernetes) GenerateRolloutYAML(
	name, namespace, image, strategy, selectorLabels string,
	replicas, minReadySeconds, progressDeadlineSeconds, cpuRequest, memoryRequest, scaleDownDelaySeconds string,
	blueGreenOptions map[string]string,
	canaryOptions map[string]string,
) (string, error) {
	// Validate required parameters
	if name == "" {
		return "", fmt.Errorf("rollout name is required")
	}
	if namespace == "" {
		namespace = "default"
	}
	if image == "" {
		return "", fmt.Errorf("container image is required")
	}
	if strategy != "canary" && strategy != "blueGreen" {
		return "", fmt.Errorf("strategy must be either 'canary' or 'blueGreen'")
	}
	if selectorLabels == "" {
		return "", fmt.Errorf("selector labels are required")
	}

	// Parse selector labels
	labels := make(map[string]string)
	for _, pair := range strings.Split(selectorLabels, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if len(labels) == 0 {
		return "", fmt.Errorf("invalid selector labels format, expected key1=value1,key2=value2")
	}

	// Begin YAML construction
	yamlBuilder := strings.Builder{}
	yamlBuilder.WriteString(fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: %s
  namespace: %s
spec:
  replicas: %s
`, name, namespace, replicas))

	// Add min ready seconds and progress deadline if specified
	if minReadySeconds != "0" {
		yamlBuilder.WriteString(fmt.Sprintf("  minReadySeconds: %s\n", minReadySeconds))
	}
	if progressDeadlineSeconds != "600" {
		yamlBuilder.WriteString(fmt.Sprintf("  progressDeadlineSeconds: %s\n", progressDeadlineSeconds))
	}

	// Add selector
	yamlBuilder.WriteString("  selector:\n    matchLabels:\n")
	for k, v := range labels {
		yamlBuilder.WriteString(fmt.Sprintf("      %s: %s\n", k, v))
	}

	// Add template
	yamlBuilder.WriteString("  template:\n    metadata:\n      labels:\n")
	for k, v := range labels {
		yamlBuilder.WriteString(fmt.Sprintf("        %s: %s\n", k, v))
	}

	// Add container spec
	yamlBuilder.WriteString(fmt.Sprintf(`    spec:
      containers:
      - name: %s
        image: %s
`, name, image))

	// Add resource requests if specified
	if cpuRequest != "" || memoryRequest != "" {
		yamlBuilder.WriteString("        resources:\n          requests:\n")
		if cpuRequest != "" {
			yamlBuilder.WriteString(fmt.Sprintf("            cpu: %s\n", cpuRequest))
		}
		if memoryRequest != "" {
			yamlBuilder.WriteString(fmt.Sprintf("            memory: %s\n", memoryRequest))
		}
	}

	// Add strategy specific configuration
	yamlBuilder.WriteString("  strategy:\n")

	if strategy == "blueGreen" {
		// Handle Blue-Green strategy
		activeService := blueGreenOptions["active_service"]
		if activeService == "" {
			return "", fmt.Errorf("active_service is required for blueGreen strategy")
		}

		previewService := blueGreenOptions["preview_service"]
		autoPromotionEnabled := blueGreenOptions["auto_promotion_enabled"]
		if autoPromotionEnabled == "" {
			autoPromotionEnabled = "false"
		}
		autoPromotionSeconds := blueGreenOptions["auto_promotion_seconds"]

		yamlBuilder.WriteString(fmt.Sprintf(`    blueGreen:
      activeService: %s
`, activeService))

		if previewService != "" {
			yamlBuilder.WriteString(fmt.Sprintf("      previewService: %s\n", previewService))
		}

		yamlBuilder.WriteString(fmt.Sprintf("      autoPromotionEnabled: %s\n", autoPromotionEnabled))

		if autoPromotionSeconds != "" {
			yamlBuilder.WriteString(fmt.Sprintf("      autoPromotionSeconds: %s\n", autoPromotionSeconds))
		}

		if scaleDownDelaySeconds != "30" {
			yamlBuilder.WriteString(fmt.Sprintf("      scaleDownDelaySeconds: %s\n", scaleDownDelaySeconds))
		}
	} else if strategy == "canary" {
		// Handle Canary strategy
		maxSurge := canaryOptions["max_surge"]
		if maxSurge == "" {
			maxSurge = "1"
		}
		maxUnavailable := canaryOptions["max_unavailable"]
		if maxUnavailable == "" {
			maxUnavailable = "0"
		}

		yamlBuilder.WriteString(fmt.Sprintf(`    canary:
      maxSurge: %s
      maxUnavailable: %s
`, maxSurge, maxUnavailable))

		// Add traffic routing if specified
		trafficRoutingProvider := canaryOptions["traffic_routing_provider"]
		if trafficRoutingProvider != "" {
			canaryService := canaryOptions["canary_service"]
			stableService := canaryOptions["stable_service"]

			if canaryService == "" || stableService == "" {
				return "", fmt.Errorf("canary_service and stable_service are required for canary strategy with traffic routing")
			}

			yamlBuilder.WriteString(fmt.Sprintf("      canaryService: %s\n", canaryService))
			yamlBuilder.WriteString(fmt.Sprintf("      stableService: %s\n", stableService))

			// Add traffic routing configuration
			yamlBuilder.WriteString("      trafficRouting:\n")

			switch trafficRoutingProvider {
			case "istio":
				yamlBuilder.WriteString("        istio:\n")
				yamlBuilder.WriteString("          virtualService:\n")
				yamlBuilder.WriteString(fmt.Sprintf("            name: %s-vsvc\n", name))
				yamlBuilder.WriteString("            routes:\n")
				yamlBuilder.WriteString("            - primary\n")
			case "nginx":
				yamlBuilder.WriteString("        nginx:\n")
				yamlBuilder.WriteString(fmt.Sprintf("          stableIngress: %s-ingress\n", name))
			case "alb":
				yamlBuilder.WriteString("        alb:\n")
				yamlBuilder.WriteString(fmt.Sprintf("          ingress: %s-ingress\n", name))
				yamlBuilder.WriteString("          servicePort: 80\n")
			case "smi":
				yamlBuilder.WriteString("        smi:\n")
				yamlBuilder.WriteString(fmt.Sprintf("          rootService: %s-root\n", name))
				yamlBuilder.WriteString(fmt.Sprintf("          trafficSplitName: %s-traffic-split\n", name))
			default:
				return "", fmt.Errorf("unsupported traffic routing provider: %s", trafficRoutingProvider)
			}
		}

		// Add steps if specified
		stepsParam := canaryOptions["steps"]
		if stepsParam != "" {
			yamlBuilder.WriteString("      steps:\n")

			stepParts := strings.Split(stepsParam, ",")
			for _, step := range stepParts {
				stepKV := strings.SplitN(step, "=", 2)
				if len(stepKV) != 2 {
					return "", fmt.Errorf("invalid step format, expected key=value")
				}

				key := strings.TrimSpace(stepKV[0])
				value := strings.TrimSpace(stepKV[1])

				switch key {
				case "setWeight":
					yamlBuilder.WriteString(fmt.Sprintf("      - setWeight: %s\n", value))
				case "pause":
					if value == "" {
						yamlBuilder.WriteString("      - pause: {}\n")
					} else {
						yamlBuilder.WriteString(fmt.Sprintf("      - pause:\n          duration: %s\n", value))
					}
				default:
					return "", fmt.Errorf("unsupported step type: %s", key)
				}
			}
		}

		// Add analysis if specified
		analysisTemplates := canaryOptions["analysis_templates"]
		if analysisTemplates != "" {
			yamlBuilder.WriteString("      analysis:\n")
			yamlBuilder.WriteString("        templates:\n")

			for _, template := range strings.Split(analysisTemplates, ",") {
				template = strings.TrimSpace(template)
				if template != "" {
					yamlBuilder.WriteString(fmt.Sprintf("        - templateName: %s\n", template))
				}
			}
		}
	}

	log.Infof("Generated Argo Rollout configuration for %s/%s", namespace, name)
	return yamlBuilder.String(), nil
}

// PauseRollout pauses an Argo Rollout to temporarily halt progression
func (k *Kubernetes) PauseRollout(ctx context.Context, name, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Construct the API endpoint for pausing rollout
	endpoint := fmt.Sprintf("/api/v1/argorollouts/rollouts/%s/%s/pause", namespace, name)

	// Make API request
	response, err := k.MakeAPIRequest("POST", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to pause rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if errMsg, ok := result["error"].(string); ok {
		return "", fmt.Errorf("API error: %s", errMsg)
	}

	// Extract the result message
	if message, ok := result["message"].(string); ok {
		log.Infof("Successfully paused rollout '%s' in namespace '%s'", name, namespace)
		return message, nil
	}

	return fmt.Sprintf("Rollout '%s' in namespace '%s' has been paused", name, namespace), nil
}

// SetRolloutImage updates the image of a container in an Argo Rollout
func (k *Kubernetes) SetRolloutImage(ctx context.Context, name, namespace, containerName, image string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	if image == "" {
		return "", fmt.Errorf("new image is required")
	}

	// Construct the API endpoint for setting rollout image
	endpoint := fmt.Sprintf("/api/v1/argorollouts/rollouts/%s/%s/image", namespace, name)

	// Prepare request body
	requestBody := map[string]interface{}{
		"image": image,
	}
	if containerName != "" {
		requestBody["container"] = containerName
	}

	// Make API request
	response, err := k.MakeAPIRequest("POST", endpoint, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to set image for rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if errMsg, ok := result["error"].(string); ok {
		return "", fmt.Errorf("API error: %s", errMsg)
	}

	// Extract the result message
	if message, ok := result["message"].(string); ok {
		if containerName != "" {
			log.Infof("Successfully updated image for container '%s' in rollout '%s' in namespace '%s' to '%s'", containerName, name, namespace, image)
		} else {
			log.Infof("Successfully updated image for rollout '%s' in namespace '%s' to '%s'", name, namespace, image)
		}
		return message, nil
	}

	if containerName != "" {
		return fmt.Sprintf("Successfully updated image for container '%s' in rollout '%s' in namespace '%s' to '%s'",
			containerName, name, namespace, image), nil
	} else {
		return fmt.Sprintf("Successfully updated image for the first container in rollout '%s' in namespace '%s' to '%s'",
			name, namespace, image), nil
	}
}
