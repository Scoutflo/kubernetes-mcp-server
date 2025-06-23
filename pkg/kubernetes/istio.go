package kubernetes

import (
	"context"
	"fmt"
	"strings"
)

// IstioStatus returns the status of Istio installation in the cluster
func (k *Kubernetes) IstioStatus(ctx context.Context) (string, error) {
	// Make API request to the Istio status endpoint
	response, err := k.MakeAPIRequest("GET", "/apis/v1/istio/status", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Istio status: %w", err)
	}
	return string(response), nil
}

// GetVirtualServices returns all virtual services in the specified namespace
func (k *Kubernetes) GetVirtualServices(ctx context.Context, namespace string) (string, error) {
	var endpoint string
	if namespace != "" {
		endpoint = fmt.Sprintf("/apis/v1/istio/virtualservices?namespace=%s", namespace)
	} else {
		endpoint = "/apis/v1/istio/virtualservices"
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get virtual services: %w", err)
	}
	return string(response), nil
}

// GetVirtualService returns a specific virtual service by name and namespace
func (k *Kubernetes) GetVirtualService(ctx context.Context, namespace, name string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	endpoint := fmt.Sprintf("/apis/v1/istio/virtualservices?namespace=%s&name=%s", namespace, name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get virtual service %s/%s: %w", namespace, name, err)
	}
	return string(response), nil
}

// GetDestinationRules returns all destination rules in the specified namespace
func (k *Kubernetes) GetDestinationRules(ctx context.Context, namespace string) (string, error) {
	var endpoint string
	if namespace != "" {
		endpoint = fmt.Sprintf("/apis/v1/istio/destinationrules?namespace=%s", namespace)
	} else {
		endpoint = "/apis/v1/istio/destinationrules"
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get destination rules: %w", err)
	}
	return string(response), nil
}

// GetDestinationRule returns a specific destination rule by name and namespace
func (k *Kubernetes) GetDestinationRule(ctx context.Context, namespace, name string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	endpoint := fmt.Sprintf("/apis/v1/istio/destinationrules?namespace=%s&name=%s", namespace, name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get destination rule %s/%s: %w", namespace, name, err)
	}
	return string(response), nil
}

// GetGateways returns all gateways in the specified namespace
func (k *Kubernetes) GetGateways(ctx context.Context, namespace string) (string, error) {
	var endpoint string
	if namespace != "" {
		endpoint = fmt.Sprintf("/apis/v1/istio/gateways?namespace=%s", namespace)
	} else {
		endpoint = "/apis/v1/istio/gateways"
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get gateways: %w", err)
	}
	return string(response), nil
}

// GetGateway returns a specific gateway by name and namespace
func (k *Kubernetes) GetGateway(ctx context.Context, namespace, name string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	endpoint := fmt.Sprintf("/apis/v1/istio/gateways?namespace=%s&name=%s", namespace, name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get gateway %s/%s: %w", namespace, name, err)
	}
	return string(response), nil
}

// GetServiceEntries returns all service entries in the specified namespace
func (k *Kubernetes) GetServiceEntries(ctx context.Context, namespace string) (string, error) {
	var endpoint string
	if namespace != "" {
		endpoint = fmt.Sprintf("/apis/v1/istio/serviceentries?namespace=%s", namespace)
	} else {
		endpoint = "/apis/v1/istio/serviceentries"
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get service entries: %w", err)
	}
	return string(response), nil
}

// GetServiceEntry returns a specific service entry by name and namespace
func (k *Kubernetes) GetServiceEntry(ctx context.Context, namespace, name string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	endpoint := fmt.Sprintf("/apis/v1/istio/serviceentries?namespace=%s&name=%s", namespace, name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get service entry %s/%s: %w", namespace, name, err)
	}
	return string(response), nil
}

// GetPeerAuthentications returns all peer authentications in the specified namespace
func (k *Kubernetes) GetPeerAuthentications(ctx context.Context, namespace string) (string, error) {
	var endpoint string
	if namespace != "" {
		endpoint = fmt.Sprintf("/apis/v1/istio/peerauthentications?namespace=%s", namespace)
	} else {
		endpoint = "/apis/v1/istio/peerauthentications"
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get peer authentications: %w", err)
	}
	return string(response), nil
}

// GetPeerAuthentication returns a specific peer authentication by name and namespace
func (k *Kubernetes) GetPeerAuthentication(ctx context.Context, namespace, name string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	endpoint := fmt.Sprintf("/apis/v1/istio/peerauthentications?namespace=%s&name=%s", namespace, name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get peer authentication %s/%s: %w", namespace, name, err)
	}
	return string(response), nil
}

// GetAuthorizationPolicies returns all authorization policies in the specified namespace
func (k *Kubernetes) GetAuthorizationPolicies(ctx context.Context, namespace string) (string, error) {
	var endpoint string
	if namespace != "" {
		endpoint = fmt.Sprintf("/apis/v1/istio/authorizationpolicies?namespace=%s", namespace)
	} else {
		endpoint = "/apis/v1/istio/authorizationpolicies"
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get authorization policies: %w", err)
	}
	return string(response), nil
}

// GetAuthorizationPolicy returns a specific authorization policy by name and namespace
func (k *Kubernetes) GetAuthorizationPolicy(ctx context.Context, namespace, name string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	endpoint := fmt.Sprintf("/apis/v1/istio/authorizationpolicies?namespace=%s&name=%s", namespace, name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get authorization policy %s/%s: %w", namespace, name, err)
	}
	return string(response), nil
}

// GetTelemetries returns all telemetries in the specified namespace
func (k *Kubernetes) GetTelemetries(ctx context.Context, namespace string) (string, error) {
	var endpoint string
	if namespace != "" {
		endpoint = fmt.Sprintf("/apis/v1/istio/telemetries?namespace=%s", namespace)
	} else {
		endpoint = "/apis/v1/istio/telemetries"
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get telemetries: %w", err)
	}
	return string(response), nil
}

// GetTelemetry returns a specific telemetry by name and namespace
func (k *Kubernetes) GetTelemetry(ctx context.Context, namespace, name string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	endpoint := fmt.Sprintf("/apis/v1/istio/telemetries?namespace=%s&name=%s", namespace, name)
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get telemetry %s/%s: %w", namespace, name, err)
	}
	return string(response), nil
}

// GetZtunnelConfig returns ztunnel configuration based on the specified parameters
func (k *Kubernetes) GetZtunnelConfig(ctx context.Context, namespace, configType, pod string) (string, error) {
	// Default namespace to istio-system if not provided
	if namespace == "" {
		namespace = "istio-system"
	}

	// Default configType to "all" if not provided
	if configType == "" {
		configType = "all"
	}

	// Build endpoint with query parameters
	endpoint := fmt.Sprintf("/apis/v1/istio/ztunnel-config?namespace=%s&configType=%s", namespace, configType)

	// Add pod parameter if provided
	if pod != "" {
		endpoint = fmt.Sprintf("%s&pod=%s", endpoint, pod)
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get ztunnel config: %w", err)
	}
	return string(response), nil
}

// GetWaypoint returns waypoint configuration/status based on the specified parameters
func (k *Kubernetes) GetWaypoint(ctx context.Context, namespace, name, action string) (string, error) {
	// Default namespace to "default" if not provided
	if namespace == "" {
		namespace = "default"
	}

	// Default name to "waypoint" if not provided
	if name == "" {
		name = "waypoint"
	}

	// Default action to "list" if not provided
	if action == "" {
		action = "list"
	}

	// Build endpoint with query parameters
	endpoint := fmt.Sprintf("/apis/v1/istio/waypoint?namespace=%s&action=%s&name=%s", namespace, action, name)

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get waypoint: %w", err)
	}
	return string(response), nil
}

// GetProxyConfig returns proxy configuration for a specific pod
func (k *Kubernetes) GetProxyConfig(ctx context.Context, podName, namespace, configType string) (string, error) {
	// podName is required
	if podName == "" {
		return "", fmt.Errorf("podName parameter is required")
	}

	// Default namespace to "default" if not provided
	if namespace == "" {
		namespace = "default"
	}

	// Default configType to "all" if not provided
	if configType == "" {
		configType = "all"
	}

	// Build endpoint with query parameters
	endpoint := fmt.Sprintf("/apis/v1/istio/proxy-config?podName=%s&namespace=%s&configType=%s&outputFormat=json", podName, namespace, configType)

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get proxy config: %w", err)
	}
	return string(response), nil
}

// GetProxyStatus returns proxy status for a specific pod or all proxies
func (k *Kubernetes) GetProxyStatus(ctx context.Context, podName, namespace string) (string, error) {
	// Build endpoint with query parameters
	endpoint := "/apis/v1/istio/proxy-status"

	var params []string

	// Add podName if provided
	if podName != "" {
		params = append(params, fmt.Sprintf("podName=%s", podName))

		// If podName is provided but namespace is not, default to "default"
		if namespace == "" {
			namespace = "default"
		}
	}

	// Add namespace if provided
	if namespace != "" {
		params = append(params, fmt.Sprintf("namespace=%s", namespace))
	}

	// Add query parameters if any
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get proxy status: %w", err)
	}
	return string(response), nil
}

// GetAnalyze performs Istio configuration analysis
func (k *Kubernetes) GetAnalyze(ctx context.Context, namespace, outputFormat, outputThreshold, failureThreshold string, allNamespaces bool) (string, error) {
	// Build endpoint with query parameters
	endpoint := "/apis/v1/istio/analyze"

	var params []string

	// Add namespace if provided
	if namespace != "" {
		params = append(params, fmt.Sprintf("namespace=%s", namespace))
	}

	// Add allNamespaces if true
	if allNamespaces {
		params = append(params, "allNamespaces=true")
	}

	// Add outputFormat if provided
	if outputFormat != "" {
		params = append(params, fmt.Sprintf("outputFormat=%s", outputFormat))
	}

	// Add outputThreshold if provided
	if outputThreshold != "" {
		params = append(params, fmt.Sprintf("outputThreshold=%s", outputThreshold))
	}

	// Add failureThreshold if provided
	if failureThreshold != "" {
		params = append(params, fmt.Sprintf("failureThreshold=%s", failureThreshold))
	}

	// Add query parameters if any
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to analyze configuration: %w", err)
	}
	return string(response), nil
}

// GetRemoteClusters lists remote clusters connected to Istiod
func (k *Kubernetes) GetRemoteClusters(ctx context.Context, revision string) (string, error) {
	// Build endpoint with query parameters
	endpoint := "/apis/v1/istio/remote-clusters"

	var params []string

	// Add revision if provided
	if revision != "" {
		params = append(params, fmt.Sprintf("revision=%s", revision))
	}

	// Add query parameters if any
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get remote clusters: %w", err)
	}
	return string(response), nil
}

// GetRequestAuthentications returns request authentications for all namespaces or a specific namespace
func (k *Kubernetes) GetRequestAuthentications(ctx context.Context, namespace string) (string, error) {
	// Build endpoint with query parameters
	endpoint := "/apis/v1/istio/requestauthentications"

	var params []string

	// Add namespace if provided
	if namespace != "" {
		params = append(params, fmt.Sprintf("namespace=%s", namespace))
	}

	// Add query parameters if any
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get request authentications: %w", err)
	}
	return string(response), nil
}

// GetRequestAuthentication returns a specific request authentication by namespace and name
func (k *Kubernetes) GetRequestAuthentication(ctx context.Context, namespace, name string) (string, error) {
	// Build endpoint with query parameters
	endpoint := "/apis/v1/istio/requestauthentications"

	var params []string

	// Add namespace and name
	if namespace != "" {
		params = append(params, fmt.Sprintf("namespace=%s", namespace))
	}
	if name != "" {
		params = append(params, fmt.Sprintf("name=%s", name))
	}

	// Add query parameters
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get request authentication: %w", err)
	}
	return string(response), nil
}

// GetWasmPlugins returns wasm plugins for all namespaces or a specific namespace
func (k *Kubernetes) GetWasmPlugins(ctx context.Context, namespace string) (string, error) {
	// Build endpoint with query parameters
	endpoint := "/apis/v1/istio/wasmplugins"

	var params []string

	// Add namespace if provided
	if namespace != "" {
		params = append(params, fmt.Sprintf("namespace=%s", namespace))
	}

	// Add query parameters if any
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
	}

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get wasm plugins: %w", err)
	}

	return string(response), nil
}

// GetWasmPlugin returns a specific wasm plugin by name and namespace
func (k *Kubernetes) GetWasmPlugin(ctx context.Context, namespace, name string) (string, error) {
	// Build endpoint with query parameters
	endpoint := "/apis/v1/istio/wasmplugins"

	var params []string

	// Add namespace and name
	params = append(params, fmt.Sprintf("namespace=%s", namespace))
	params = append(params, fmt.Sprintf("name=%s", name))

	// Add query parameters
	endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))

	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get wasm plugin %s/%s: %w", namespace, name, err)
	}

	return string(response), nil
}
