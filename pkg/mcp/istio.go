package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initIstio() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("istio_status",
			mcp.WithDescription("Verify the operational status of Istio components, such as Istiod and proxies, to ensure service mesh reliability. This tool confirms setup integrity for monitoring and incident response."),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioStatus},
		{Tool: mcp.NewTool("istio_get_virtual_services",
			mcp.WithDescription("Retrieve Istio virtual service configurations to inspect traffic routing rules. This tool diagnoses routing issues impacting service communication during troubleshooting."),
			mcp.WithString("name",
				mcp.Description("Name of the specific virtual service to get (optional, if not provided will list all virtual services)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get virtual services from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetVirtualServices},
		{Tool: mcp.NewTool("istio_get_destination_rules",
			mcp.WithDescription("Retrieve Istio destination rule configurations to verify traffic policies like load balancing or TLS. This tool supports troubleshooting by identifying policy misconfigurations."),
			mcp.WithString("name",
				mcp.Description("Name of the specific destination rule to get (optional, if not provided will list all destination rules)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get destination rules from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetDestinationRules},
		{Tool: mcp.NewTool("istio_get_gateways",
			mcp.WithDescription("Retrieve Istio gateway configurations to ensure proper ingress and egress traffic management. This tool diagnoses issues like port misconfigurations affecting access."),
			mcp.WithString("name",
				mcp.Description("Name of the specific gateway to get (optional, if not provided will list all gateways)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get gateways from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetGateways},
		{Tool: mcp.NewTool("istio_get_service_entries",
			mcp.WithDescription("Retrieve Istio service entry configurations to verify external service integration. This tool supports troubleshooting by identifying misconfigured external endpoints."),
			mcp.WithString("name",
				mcp.Description("Name of the specific service entry to get (optional, if not provided will list all service entries)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get service entries from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetServiceEntries},
		{Tool: mcp.NewTool("istio_get_peer_authentications",
			mcp.WithDescription("Retrieve Istio peer authentication policies to confirm mutual TLS enforcement. This tool diagnoses authentication issues impacting secure communication during troubleshooting."),
			mcp.WithString("name",
				mcp.Description("Name of the specific peer authentication to get (optional, if not provided will list all peer authentications)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get peer authentications from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetPeerAuthentications},
		{Tool: mcp.NewTool("istio_get_request_authentications",
			mcp.WithDescription("Retrieve Istio request authentication configurations to verify authorization settings like JWT validation. This tool supports troubleshooting by resolving access control issues."),
			mcp.WithString("name",
				mcp.Description("Name of the specific request authentication to get (optional, if not provided will list all request authentications)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get request authentications from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetRequestAuthentications},
		{Tool: mcp.NewTool("istio_get_wasm_plugins",
			mcp.WithDescription("Retrieve Istio WebAssembly plugin configurations to verify custom extensions. This tool diagnoses issues like plugin misconfigurations affecting service behavior during troubleshooting."),
			mcp.WithString("name",
				mcp.Description("Name of the specific wasm plugin to get (optional, if not provided will list all wasm plugins)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get wasm plugins from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetWasmPlugins},
		{Tool: mcp.NewTool("istio_get_authorization_policies",
			mcp.WithDescription("Retrieve Istio authorization policies to confirm access control rules. This tool supports troubleshooting by identifying restrictive policies impacting legitimate traffic."),
			mcp.WithString("name",
				mcp.Description("Name of the specific authorization policy to get (optional, if not provided will list all authorization policies)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get authorization policies from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetAuthorizationPolicies},
		{Tool: mcp.NewTool("istio_get_telemetries",
			mcp.WithDescription("Retrieve Istio telemetry configurations to ensure proper metric and log collection. This tool supports historical analysis by diagnosing issues like missing performance metrics."),
			mcp.WithString("name",
				mcp.Description("Name of the specific telemetry to get (optional, if not provided will list all telemetries)"),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace to get telemetries from (optional, if not provided will get from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.istioGetTelemetries},
		{Tool: mcp.NewTool("istio_get_ztunnel_config",
			mcp.WithDescription("Retrieve Istio ztunnel configurations to verify secure communication settings. This tool diagnoses connectivity issues like encryption misconfigurations during troubleshooting."),
			mcp.WithString("ns",
				mcp.Description("The namespace of the pod to get ztunnel configuration for (optional, defaults to istio-system)"),
			),
			mcp.WithString("config_type",
				mcp.Description("The type of configuration to get, the allowed values are: all, workload, service, policy, certificate, connections, log (optional, defaults to all)"),
			),
			mcp.WithString("pod",
				mcp.Description("Pod name (optional, if not provided gets config from all pods)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.ztunnelConfig},
		{Tool: mcp.NewTool("istio_get_waypoint",
			mcp.WithDescription("Retrieve Istio waypoint configurations to inspect or manage traffic settings. This tool supports testing routing changes for incident resolution or deployment validation."),
			mcp.WithString("name",
				mcp.Description("Name of the waypoint to get status for"),
				mcp.Required(),
			),
			mcp.WithString("ns",
				mcp.Description("Namespace of the waypoint to get status for"),
				mcp.Required(),
			),
			mcp.WithString("action",
				mcp.Description("Waypoint action: list, status, generate, apply, delete (optional, defaults to status)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.waypoint},
		{Tool: mcp.NewTool("istio_get_proxy_config",
			mcp.WithDescription("Retrieve Envoy proxy configurations for a specific pod to inspect traffic handling. This tool diagnoses issues like misconfigured filters affecting requests during troubleshooting."),
			mcp.WithString("pod_name",
				mcp.Description("The name of the pod to get proxy configuration for"),
				mcp.Required(),
			),
			mcp.WithString("ns",
				mcp.Description("The namespace of the pod to get proxy configuration for (optional, defaults to default)"),
			),
			mcp.WithString("config_type",
				mcp.Description("The type of configuration to get, the allowed values are: all, bootstrap, cluster, listener, route, endpoint, secret, log (optional, defaults to all)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.proxyConfig},
		{Tool: mcp.NewTool("istio_get_proxy_status",
			mcp.WithDescription("Retrieve the synchronization status between Istiod and Envoy proxies to verify configuration alignment. This tool diagnoses delays or failures impacting service mesh operations in real-time."),
			mcp.WithString("pod_name",
				mcp.Description("The name of the pod to get Envoy proxy status for (optional, if not provided gets status for all pods)"),
			),
			mcp.WithString("ns",
				mcp.Description("The namespace of the pod to get Envoy proxy status for (optional, if not provided gets status from all namespaces)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.proxyStatus},
		{Tool: mcp.NewTool("istio_analyze_cluster_configuration",
			mcp.WithDescription("Analyze Istio cluster configurations to identify misconfigurations or gaps. This tool generates reports to support historical root cause analysis of service mesh issues."),
			mcp.WithString("namespace",
				mcp.Description("Namespace to analyze (optional, analyzes all namespaces if not specified)"),
			),
			mcp.WithString("output_format",
				mcp.Description("Output format: log, json, yaml (optional, defaults to json)"),
			),
			mcp.WithString("output_threshold",
				mcp.Description("Output threshold: Info, Warning, Error (optional, defaults to Info)"),
			),
			mcp.WithString("failure_threshold",
				mcp.Description("Failure threshold: Info, Warning, Error (optional, defaults to Error)"),
			),
			mcp.WithBoolean("all_namespaces",
				mcp.Description("Analyze all namespaces (optional, defaults to false)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.analyzeClusterConfiguration},
		{Tool: mcp.NewTool("istio_get_remote_clusters",
			mcp.WithDescription("List remote clusters connected to the Istio control plane to verify multi-cluster setup integrity. This tool supports troubleshooting by identifying reachability issues across clusters."),
			mcp.WithString("revision",
				mcp.Description("Control plane revision to check (optional, defaults to default)"),
			),
			mcp.WithString("k8surl", mcp.Description("Kubernetes API server URL"), mcp.Required()),
			mcp.WithString("k8stoken", mcp.Description("Kubernetes API server authentication token"), mcp.Required()),
		), Handler: s.remoteClusters},
	}
}

// istioStatus handles the istio_status tool request
func (s *Server) istioStatus(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_status failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: istio_status - checking Istio installation status - got called by session id: %s", sessionID)

	ret, err := k.IstioStatus(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: istio_status failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Istio status: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: istio_status completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetVirtualServices handles the istio_get_virtual_services tool request
func (s *Server) istioGetVirtualServices(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_virtual_services failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific virtual service
		klog.V(1).Infof("Tool: istio_get_virtual_services - getting virtual service: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetVirtualService(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_virtual_services failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get virtual service %s: %v", name, err)), nil
		}
	} else {
		// List all virtual services
		klog.V(1).Infof("Tool: istio_get_virtual_services - getting virtual services in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetVirtualServices(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_virtual_services failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get virtual services: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_virtual_services completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetDestinationRules handles the istio_get_destination_rules tool request
func (s *Server) istioGetDestinationRules(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_destination_rules failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific destination rule
		klog.V(1).Infof("Tool: istio_get_destination_rules - getting destination rule: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetDestinationRule(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_destination_rules failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get destination rule %s: %v", name, err)), nil
		}
	} else {
		// List all destination rules
		klog.V(1).Infof("Tool: istio_get_destination_rules - getting destination rules in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetDestinationRules(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_destination_rules failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get destination rules: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_destination_rules completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetGateways handles the istio_get_gateways tool request
func (s *Server) istioGetGateways(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_gateways failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific gateway
		klog.V(1).Infof("Tool: istio_get_gateways - getting gateway: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetGateway(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_gateways failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get gateway %s: %v", name, err)), nil
		}
	} else {
		// List all gateways
		klog.V(1).Infof("Tool: istio_get_gateways - getting gateways in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetGateways(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_gateways failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get gateways: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_gateways completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetServiceEntries handles the istio_get_service_entries tool request
func (s *Server) istioGetServiceEntries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_service_entries failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific service entry
		klog.V(1).Infof("Tool: istio_get_service_entries - getting service entry: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetServiceEntry(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_service_entries failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get service entry %s: %v", name, err)), nil
		}
	} else {
		// List all service entries
		klog.V(1).Infof("Tool: istio_get_service_entries - getting service entries in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetServiceEntries(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_service_entries failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get service entries: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_service_entries completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetPeerAuthentications handles the istio_get_peer_authentications tool request
func (s *Server) istioGetPeerAuthentications(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_peer_authentications failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific peer authentication
		klog.V(1).Infof("Tool: istio_get_peer_authentications - getting peer authentication: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetPeerAuthentication(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_peer_authentications failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get peer authentication %s: %v", name, err)), nil
		}
	} else {
		// List all peer authentications
		klog.V(1).Infof("Tool: istio_get_peer_authentications - getting peer authentications in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetPeerAuthentications(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_peer_authentications failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get peer authentications: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_peer_authentications completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetRequestAuthentications handles the istio_get_request_authentications tool request
func (s *Server) istioGetRequestAuthentications(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_request_authentications failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific request authentication
		klog.V(1).Infof("Tool: istio_get_request_authentications - getting request authentication: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetRequestAuthentication(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_request_authentications failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get request authentication %s: %v", name, err)), nil
		}
	} else {
		// List all request authentications
		klog.V(1).Infof("Tool: istio_get_request_authentications - getting request authentications in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetRequestAuthentications(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_request_authentications failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get request authentications: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_request_authentications completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetWasmPlugins handles the istio_get_wasm_plugins tool request
func (s *Server) istioGetWasmPlugins(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_wasm_plugins failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific wasm plugin
		klog.V(1).Infof("Tool: istio_get_wasm_plugins - getting wasm plugin: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetWasmPlugin(ctx, namespace, name)
		duration = time.Since(start)
	} else {
		// List wasm plugins
		klog.V(1).Infof("Tool: istio_get_wasm_plugins - listing wasm plugins in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetWasmPlugins(ctx, namespace)
		duration = time.Since(start)
	}

	if err != nil {
		klog.Errorf("Tool call: istio_get_wasm_plugins failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get wasm plugins: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: istio_get_wasm_plugins completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetAuthorizationPolicies handles the istio_get_authorization_policies tool request
func (s *Server) istioGetAuthorizationPolicies(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_authorization_policies failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific authorization policy
		klog.V(1).Infof("Tool: istio_get_authorization_policies - getting authorization policy: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetAuthorizationPolicy(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_authorization_policies failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get authorization policy %s: %v", name, err)), nil
		}
	} else {
		// List all authorization policies
		klog.V(1).Infof("Tool: istio_get_authorization_policies - getting authorization policies in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetAuthorizationPolicies(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_authorization_policies failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get authorization policies: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_authorization_policies completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// istioGetTelemetries handles the istio_get_telemetries tool request
func (s *Server) istioGetTelemetries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: istio_get_telemetries failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("namespace", "")

	var ret string
	var duration time.Duration

	if name != "" {
		// Get specific telemetry
		klog.V(1).Infof("Tool: istio_get_telemetries - getting telemetry: %s in namespace: %s - got called by session id: %s", name, namespace, sessionID)
		ret, err = k.GetTelemetry(ctx, namespace, name)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_telemetries failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get telemetry %s: %v", name, err)), nil
		}
	} else {
		// List all telemetries
		klog.V(1).Infof("Tool: istio_get_telemetries - getting telemetries in namespace: %s - got called by session id: %s", namespace, sessionID)
		ret, err = k.GetTelemetries(ctx, namespace)
		duration = time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: istio_get_telemetries failed after %v: %v by session id: %s", duration, err, sessionID)
			return NewTextResult("", fmt.Errorf("failed to get telemetries: %v", err)), nil
		}
	}

	klog.V(1).Infof("Tool call: istio_get_telemetries completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// ztunnelConfig handles the ztunnel_config tool request
func (s *Server) ztunnelConfig(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: ztunnel_config failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	namespace := ctr.GetString("ns", "")
	configType := ctr.GetString("config_type", "")
	pod := ctr.GetString("pod", "")

	klog.V(1).Infof("Tool: ztunnel_config - getting ztunnel config in namespace: %s, configType: %s, pod: %s - got called by session id: %s", namespace, configType, pod, sessionID)

	ret, err := k.GetZtunnelConfig(ctx, namespace, configType, pod)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: ztunnel_config failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get ztunnel config: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: ztunnel_config completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// waypoint handles the waypoint tool request
func (s *Server) waypoint(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: waypoint failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	name := ctr.GetString("name", "")
	namespace := ctr.GetString("ns", "")
	action := ctr.GetString("action", "status")

	klog.V(1).Infof("Tool: waypoint - getting waypoint: %s in namespace: %s with action: %s - got called by session id: %s", name, namespace, action, sessionID)

	if name == "" {
		klog.Errorf("Tool call: waypoint failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("name parameter is required")), nil
	}

	if namespace == "" {
		klog.Errorf("Tool call: waypoint failed after %v: missing ns parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("ns parameter is required")), nil
	}

	ret, err := k.GetWaypoint(ctx, namespace, name, action)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: waypoint failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get waypoint: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: waypoint completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// proxyConfig handles the proxy_config tool request
func (s *Server) proxyConfig(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: proxy_config failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	podName := ctr.GetString("pod_name", "")
	namespace := ctr.GetString("ns", "")
	configType := ctr.GetString("config_type", "")

	klog.V(1).Infof("Tool: proxy_config - getting proxy config for pod: %s in namespace: %s with configType: %s - got called by session id: %s", podName, namespace, configType, sessionID)

	if podName == "" {
		klog.Errorf("Tool call: proxy_config failed after %v: missing pod_name parameter", time.Since(start))
		return NewTextResult("", fmt.Errorf("pod_name parameter is required")), nil
	}

	ret, err := k.GetProxyConfig(ctx, podName, namespace, configType)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: proxy_config failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get proxy config: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: proxy_config completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// proxyStatus handles the proxy_status tool request
func (s *Server) proxyStatus(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: proxy_status failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	podName := ctr.GetString("pod_name", "")
	namespace := ctr.GetString("ns", "")

	klog.V(1).Infof("Tool: proxy_status - getting proxy status for pod: %s in namespace: %s - got called by session id: %s", podName, namespace, sessionID)

	ret, err := k.GetProxyStatus(ctx, podName, namespace)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: proxy_status failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get proxy status: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: proxy_status completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// analyzeClusterConfiguration handles the analyze_cluster_configuration tool request
func (s *Server) analyzeClusterConfiguration(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: analyze_cluster_configuration failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	namespace := ctr.GetString("namespace", "")
	outputFormat := ctr.GetString("output_format", "json")
	outputThreshold := ctr.GetString("output_threshold", "Info")
	failureThreshold := ctr.GetString("failure_threshold", "Error")
	allNamespaces := ctr.GetBool("all_namespaces", false)

	klog.V(1).Infof("Tool: analyze_cluster_configuration - analyzing configuration for namespace: %s, all_namespaces: %v - got called by session id: %s", namespace, allNamespaces, sessionID)

	ret, err := k.GetAnalyze(ctx, namespace, outputFormat, outputThreshold, failureThreshold, allNamespaces)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: analyze_cluster_configuration failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to analyze cluster configuration: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: analyze_cluster_configuration completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// remoteClusters handles the remote_clusters tool request
func (s *Server) remoteClusters(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: remote_clusters failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	revision := ctr.GetString("revision", "default")

	klog.V(1).Infof("Tool: remote_clusters - getting remote clusters for revision: %s - got called by session id: %s", revision, sessionID)

	ret, err := k.GetRemoteClusters(ctx, revision)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: remote_clusters failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get remote clusters: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: remote_clusters completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}
