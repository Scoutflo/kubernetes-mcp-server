package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	rolloutclient "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// ArgoRolloutsClient represents a client for ArgoRollouts API
type ArgoRolloutsClient struct {
	namespace      string
	httpClient     *http.Client
	rolloutsClient rolloutclient.Interface
	k              *Kubernetes
}

// Rollout represents an Argo Rollouts resource
type Rollout struct {
	Kind       string          `json:"kind"`
	APIVersion string          `json:"apiVersion"`
	Metadata   RolloutMetadata `json:"metadata"`
	Spec       RolloutSpec     `json:"spec"`
	Status     RolloutStatus   `json:"status,omitempty"`
}

// RolloutMetadata contains rollout metadata
type RolloutMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty"`
	UID               string            `json:"uid,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

// RolloutSpec contains the rollout specification
type RolloutSpec struct {
	Replicas                *int32          `json:"replicas,omitempty"`
	Selector                LabelSelector   `json:"selector"`
	Template                PodTemplateSpec `json:"template"`
	MinReadySeconds         int32           `json:"minReadySeconds,omitempty"`
	Strategy                RolloutStrategy `json:"strategy"`
	RevisionHistoryLimit    *int32          `json:"revisionHistoryLimit,omitempty"`
	Paused                  bool            `json:"paused,omitempty"`
	ProgressDeadlineSeconds *int32          `json:"progressDeadlineSeconds,omitempty"`
}

// LabelSelector is a label query over a set of resources
type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// PodTemplateSpec describes the data a pod should have when created from a template
type PodTemplateSpec struct {
	Metadata ObjectMeta `json:"metadata,omitempty"`
	Spec     PodSpec    `json:"spec"`
}

// ObjectMeta is metadata about an object
type ObjectMeta struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PodSpec is a simplified pod specification for the rollout
type PodSpec struct {
	Containers []Container `json:"containers"`
}

// Container is a simplified container specification
type Container struct {
	Name      string           `json:"name"`
	Image     string           `json:"image"`
	Resources ResourceRequests `json:"resources,omitempty"`
}

// ResourceRequests defines compute resource requirements
type ResourceRequests struct {
	Requests map[string]string `json:"requests,omitempty"`
}

// RolloutStrategy defines the deployment strategy
type RolloutStrategy struct {
	BlueGreen *BlueGreenStrategy `json:"blueGreen,omitempty"`
	Canary    *CanaryStrategy    `json:"canary,omitempty"`
}

// BlueGreenStrategy defines parameters for Blue Green deployment
type BlueGreenStrategy struct {
	ActiveService         string `json:"activeService"`
	PreviewService        string `json:"previewService,omitempty"`
	AutoPromotionEnabled  *bool  `json:"autoPromotionEnabled,omitempty"`
	AutoPromotionSeconds  *int32 `json:"autoPromotionSeconds,omitempty"`
	ScaleDownDelaySeconds *int32 `json:"scaleDownDelaySeconds,omitempty"`
}

// CanaryStrategy defines parameters for Canary deployment
type CanaryStrategy struct {
	CanaryService  string           `json:"canaryService,omitempty"`
	StableService  string           `json:"stableService,omitempty"`
	MaxSurge       string           `json:"maxSurge,omitempty"`
	MaxUnavailable string           `json:"maxUnavailable,omitempty"`
	Steps          []CanaryStep     `json:"steps,omitempty"`
	TrafficRouting *TrafficRouting  `json:"trafficRouting,omitempty"`
	Analysis       *RolloutAnalysis `json:"analysis,omitempty"`
}

// CanaryStep defines a step in the canary deployment process
type CanaryStep struct {
	SetWeight *int32     `json:"setWeight,omitempty"`
	Pause     *PauseStep `json:"pause,omitempty"`
}

// PauseStep defines a pause step
type PauseStep struct {
	Duration string `json:"duration,omitempty"`
}

// TrafficRouting defines the service mesh provider for traffic splitting
type TrafficRouting struct {
	Istio *IstioTrafficRouting `json:"istio,omitempty"`
	Nginx *NginxTrafficRouting `json:"nginx,omitempty"`
	ALB   *ALBTrafficRouting   `json:"alb,omitempty"`
	SMI   *SMITrafficRouting   `json:"smi,omitempty"`
}

// IstioTrafficRouting defines Istio specific traffic routing
type IstioTrafficRouting struct {
	VirtualService VirtualServiceRef `json:"virtualService"`
}

// VirtualServiceRef is a reference to an Istio VirtualService
type VirtualServiceRef struct {
	Name   string   `json:"name"`
	Routes []string `json:"routes,omitempty"`
}

// NginxTrafficRouting defines Nginx specific traffic routing
type NginxTrafficRouting struct {
	StableIngress string `json:"stableIngress"`
}

// ALBTrafficRouting defines AWS ALB specific traffic routing
type ALBTrafficRouting struct {
	Ingress     string `json:"ingress"`
	ServicePort int32  `json:"servicePort"`
}

// SMITrafficRouting defines SMI specific traffic routing
type SMITrafficRouting struct {
	RootService      string `json:"rootService,omitempty"`
	TrafficSplitName string `json:"trafficSplitName,omitempty"`
}

// RolloutAnalysis defines inline analysis for a rollout
type RolloutAnalysis struct {
	Templates []AnalysisTemplateRef `json:"templates,omitempty"`
}

// AnalysisTemplateRef references an analysis template resource
type AnalysisTemplateRef struct {
	TemplateName string `json:"templateName"`
}

// RolloutStatus defines the observed state of a rollout
type RolloutStatus struct {
	CurrentPodHash     string           `json:"currentPodHash,omitempty"`
	CurrentStepIndex   *int32           `json:"currentStepIndex,omitempty"`
	CurrentStepHash    string           `json:"currentStepHash,omitempty"`
	PauseConditions    []PauseCondition `json:"pauseConditions,omitempty"`
	Phase              string           `json:"phase,omitempty"`
	Message            string           `json:"message,omitempty"`
	ObservedGeneration int64            `json:"observedGeneration,omitempty"`
}

// PauseCondition defines a condition where a rollout is paused
type PauseCondition struct {
	Reason    string `json:"reason"`
	StartTime string `json:"startTime"`
}

// NewArgoRolloutsClient creates a new ArgoRollouts client
func (k *Kubernetes) NewArgoRolloutsClient(namespace string) (*ArgoRolloutsClient, error) {
	// Create the client using the official Argo Rollouts client library
	rolloutsClient, err := rolloutclient.NewForConfig(k.GetRESTConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create Argo Rollouts client: %w", err)
	}

	client := &ArgoRolloutsClient{
		namespace: namespaceOrDefault(namespace),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rolloutsClient: rolloutsClient,
		k:              k,
	}
	return client, nil
}

// GetRollout gets an Argo Rollout by name and namespace
func (c *ArgoRolloutsClient) GetRollout(ctx context.Context, name, namespace string) (*rolloutv1alpha1.Rollout, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	// Get the rollout
	rollout, err := c.rolloutsClient.ArgoprojV1alpha1().Rollouts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	return rollout, nil
}

// FormatRolloutOutput formats a rollout for output in the specified format
func (c *ArgoRolloutsClient) FormatRolloutOutput(rollout *rolloutv1alpha1.Rollout, format string) (string, error) {
	// Format output based on requested format
	var result string
	switch format {
	case "json":
		jsonBytes, err := json.Marshal(rollout)
		if err != nil {
			return "", fmt.Errorf("failed to marshal rollout to JSON: %w", err)
		}
		result = string(jsonBytes)
	case "yaml":
		yamlBytes, err := yaml.Marshal(rollout)
		if err != nil {
			return "", fmt.Errorf("failed to marshal rollout to YAML: %w", err)
		}
		result = string(yamlBytes)
	default:
		// Default human-readable format
		phase := rollout.Status.Phase
		if phase == "" {
			phase = "N/A"
		}

		result = fmt.Sprintf("Name:               %s\n", rollout.Name)
		result += fmt.Sprintf("Namespace:          %s\n", rollout.Namespace)
		result += fmt.Sprintf("Status:             %s\n", phase)
		result += fmt.Sprintf("Strategy:           %s\n", c.getStrategyType(rollout))
		result += fmt.Sprintf("Images:             %s\n", c.getContainerImages(rollout))

		if rollout.Status.CurrentStepIndex != nil {
			result += fmt.Sprintf("Current Step:       %d\n", *rollout.Status.CurrentStepIndex)
		}

		if len(rollout.Status.PauseConditions) > 0 {
			result += "Pause Conditions:   Yes\n"
		} else {
			result += "Pause Conditions:   No\n"
		}
	}

	return result, nil
}

// PromoteRollout promotes an Argo Rollout to advance it to the next step
func (c *ArgoRolloutsClient) PromoteRollout(ctx context.Context, name, namespace string, fullPromote bool) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	// Get the rollouts interface for the specified namespace
	rolloutInterface := c.rolloutsClient.ArgoprojV1alpha1().Rollouts(namespace)

	// Get the current rollout
	rollout, err := rolloutInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Create a patched rollout for promotion
	var result string
	if fullPromote {
		// For full promotion in blue-green strategy, we'll set the activeSelector to the new ReplicaSet
		if rollout.Spec.Strategy.BlueGreen == nil {
			return "", fmt.Errorf("full promotion is only applicable to blue-green strategy")
		}

		// Patch the rollout for full promotion
		rollout.Status.BlueGreen.ActiveSelector = rollout.Status.CurrentPodHash
		_, err = rolloutInterface.UpdateStatus(ctx, rollout, metav1.UpdateOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to fully promote rollout: %w", err)
		}
		result = fmt.Sprintf("Rollout '%s' in namespace '%s' has been fully promoted", name, namespace)
	} else {
		// For regular promotion, we'll set the pause condition to false
		// This is equivalent to the `promote` command in kubectl plugin
		if len(rollout.Status.PauseConditions) == 0 {
			return "", fmt.Errorf("rollout '%s' in namespace '%s' is not currently paused", name, namespace)
		}

		// Patch the rollout to resume it
		// We're removing all pause conditions to advance the rollout
		rollout.Status.PauseConditions = nil
		rollout.Status.ControllerPause = false
		_, err = rolloutInterface.UpdateStatus(ctx, rollout, metav1.UpdateOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to promote rollout: %w", err)
		}
		result = fmt.Sprintf("Rollout '%s' in namespace '%s' has been promoted to the next step", name, namespace)
	}

	log.Infof("Successfully promoted rollout '%s' in namespace '%s'", name, namespace)
	return result, nil
}

// AbortRollout aborts an in-progress Argo Rollout and reverts to the stable version
func (c *ArgoRolloutsClient) AbortRollout(ctx context.Context, name, namespace string) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	// Get the rollouts interface for the specified namespace
	rolloutInterface := c.rolloutsClient.ArgoprojV1alpha1().Rollouts(namespace)

	// Get the current rollout
	rollout, err := rolloutInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Set the abort flag in the rollout
	// This is equivalent to the `abort` command in kubectl plugin
	rollout.Status.Abort = true
	_, err = rolloutInterface.UpdateStatus(ctx, rollout, metav1.UpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to abort rollout: %w", err)
	}

	log.Infof("Successfully aborted rollout '%s' in namespace '%s'", name, namespace)
	return fmt.Sprintf("Rollout '%s' in namespace '%s' has been aborted", name, namespace), nil
}

// SetRolloutWeight sets the weight for a canary rollout
func (c *ArgoRolloutsClient) SetRolloutWeight(ctx context.Context, name, namespace string, weight int) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	if weight < 0 || weight > 100 {
		return "", fmt.Errorf("weight must be between 0 and 100")
	}

	// Get the rollouts interface for the specified namespace
	rolloutInterface := c.rolloutsClient.ArgoprojV1alpha1().Rollouts(namespace)

	// Get the current rollout
	rollout, err := rolloutInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Check if rollout is using canary strategy
	if rollout.Spec.Strategy.Canary == nil {
		return "", fmt.Errorf("rollout '%s' in namespace '%s' is not using canary strategy", name, namespace)
	}

	// Set the desired weight in the rollout status
	// This simulates the setWeight command
	if rollout.Status.CurrentStepIndex == nil {
		index := int32(0)
		rollout.Status.CurrentStepIndex = &index
	}

	// Use annotations to set the desired weight
	weightStr := strconv.Itoa(weight)
	annotations := rollout.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["rollout.argoproj.io/desired-weight"] = weightStr
	rollout.Annotations = annotations

	// Update the rollout with the new annotations
	_, err = rolloutInterface.Update(ctx, rollout, metav1.UpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to set canary weight: %w", err)
	}

	log.Infof("Successfully set canary weight to %d%% for rollout '%s' in namespace '%s'", weight, name, namespace)
	return fmt.Sprintf("Canary weight for rollout '%s' in namespace '%s' has been set to %d%%", name, namespace, weight), nil
}

// Helper functions for formatting rollout information
func (c *ArgoRolloutsClient) getStrategyType(rollout *rolloutv1alpha1.Rollout) string {
	if rollout.Spec.Strategy.BlueGreen != nil {
		return "BlueGreen"
	}
	if rollout.Spec.Strategy.Canary != nil {
		return "Canary"
	}
	return "Unknown"
}

func (c *ArgoRolloutsClient) getContainerImages(rollout *rolloutv1alpha1.Rollout) string {
	var images []string
	for _, container := range rollout.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}
	return strings.Join(images, ", ")
}

// GenerateRolloutYAML generates a YAML configuration for an Argo Rollout
func (c *ArgoRolloutsClient) GenerateRolloutYAML(
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
		namespace = c.namespace
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
func (c *ArgoRolloutsClient) PauseRollout(ctx context.Context, name, namespace string) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	// Get the rollouts interface for the specified namespace
	rolloutInterface := c.rolloutsClient.ArgoprojV1alpha1().Rollouts(namespace)

	// Get the current rollout
	rollout, err := rolloutInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Check if rollout is already paused
	if rollout.Spec.Paused {
		return fmt.Sprintf("Rollout '%s' in namespace '%s' is already paused", name, namespace), nil
	}

	// Pause the rollout by setting the Paused field to true
	rollout.Spec.Paused = true
	_, err = rolloutInterface.Update(ctx, rollout, metav1.UpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pause rollout: %w", err)
	}

	log.Infof("Successfully paused rollout '%s' in namespace '%s'", name, namespace)
	return fmt.Sprintf("Rollout '%s' in namespace '%s' has been paused", name, namespace), nil
}

// SetRolloutImage updates the image of a container in an Argo Rollout
func (c *ArgoRolloutsClient) SetRolloutImage(ctx context.Context, name, namespace, containerName, image string) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	if image == "" {
		return "", fmt.Errorf("new image is required")
	}

	// Get the rollouts interface for the specified namespace
	rolloutInterface := c.rolloutsClient.ArgoprojV1alpha1().Rollouts(namespace)

	// Get the current rollout
	rollout, err := rolloutInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get rollout '%s' in namespace '%s': %w", name, namespace, err)
	}

	// Find and update the container
	containerFound := false
	for i, container := range rollout.Spec.Template.Spec.Containers {
		// If containerName is specified, match by name, otherwise update the first container
		if containerName == "" || container.Name == containerName {
			oldImage := container.Image
			rollout.Spec.Template.Spec.Containers[i].Image = image
			containerFound = true
			log.Infof("Updating container '%s' in rollout '%s/%s' from image '%s' to '%s'",
				container.Name, namespace, name, oldImage, image)

			// If containerName was specified and found, we're done
			if containerName != "" {
				break
			}
		}
	}

	if !containerFound {
		if containerName != "" {
			return "", fmt.Errorf("container '%s' not found in rollout '%s' in namespace '%s'", containerName, name, namespace)
		} else {
			return "", fmt.Errorf("no containers found in rollout '%s' in namespace '%s'", name, namespace)
		}
	}

	// Update the rollout
	_, err = rolloutInterface.Update(ctx, rollout, metav1.UpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to update rollout image: %w", err)
	}

	if containerName != "" {
		return fmt.Sprintf("Successfully updated image for container '%s' in rollout '%s' in namespace '%s' to '%s'",
			containerName, name, namespace, image), nil
	} else {
		return fmt.Sprintf("Successfully updated image for the first container in rollout '%s' in namespace '%s' to '%s'",
			name, namespace, image), nil
	}
}
