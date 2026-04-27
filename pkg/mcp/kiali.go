package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initKiali() []server.ServerTool {
	return []server.ServerTool{
		{Tool: WithMeta(
			mcp.NewTool("kiali_health_check",
				mcp.WithDescription("Check whether the Kiali service is accessible and responding. Call first before any other kiali_* tool when Kiali-based investigation is required — if this returns unhealthy, all other Kiali tools will fail and you should fall back to raw Istio tools (get_virtual_services, get_proxy_config, etc.)."),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiHealthCheck},
		{Tool: WithMeta(
			mcp.NewTool("kiali_namespace_validations",
				mcp.WithDescription("Validate all Istio configuration objects in a namespace and surface semantic or runtime errors. Scans VirtualServices, DestinationRules, and auth policies for invalid hosts, conflicting rules, and misconfigured weights. Call this before diving into individual istio get_* tools — it narrows which specific resources need inspection and avoids broad config scanning. Use as the first Istio triage step when traffic is misbehaving but there is no obvious error in events_list."),
				mcp.WithString("namespace", mcp.Description("Namespace to get validations for"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiNamespaceValidations},
		{Tool: WithMeta(
			mcp.NewTool("kiali_namespace_tls",
				mcp.WithDescription("Audit the effective mTLS status (STRICT/PERMISSIVE/DISABLED) for all workloads in a namespace. Synthesizes PeerAuthentication and DestinationRule state into a single view. Use when diagnosing mTLS-related connection failures — call this before get_peer_authentications or get_destination_rules to confirm whether the issue is a mode mismatch rather than a config syntax error. Also use for compliance checks requiring all traffic to be STRICT."),
				mcp.WithString("namespace", mcp.Description("Namespace to get TLS status for"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiNamespaceTLS},
		{Tool: WithMeta(
			mcp.NewTool("kiali_clusters_services",
				mcp.WithDescription("List all services across clusters with their Istio sidecar injection status. Returns service metadata, labels, and whether each service participates in the mesh. Use to discover service names and namespaces when unknown — required as a prerequisite before kiali_service_metrics or kiali_service_metrics when the service name is not established. Also use to identify services missing sidecar injection that should be in the mesh."),
				mcp.WithString("namespace", mcp.Description("Filter by specific namespace (optional)")),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiClustersServices},
		{Tool: WithMeta(
			mcp.NewTool("kiali_istio_config",
				mcp.WithDescription("List all Istio configuration objects (VirtualServices, DestinationRules, Gateways, ServiceEntries) across the cluster with normalized specs. Use for cluster-wide config inventory or drift detection. Prefer kiali_namespace_validations when you already know which namespace has an issue — it surfaces errors rather than just listing objects. Use kiali_istio_config when the namespace is unknown or when you need to compare config across multiple namespaces."),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiIstioConfig},
		{Tool: WithMeta(
			mcp.NewTool("kiali_tracing_info",
				mcp.WithDescription("Check whether the distributed tracing backend (Jaeger/Tempo) is connected to Kiali and available. Reports tracing system status, endpoint reachability, and configuration errors. Call before attempting trace-based investigation to confirm trace data is accessible — if tracing is unavailable, fall back to kiali_service_metrics and kiali_namespace_metrics for latency analysis."),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiTracingInfo},
		{Tool: WithMeta(
			mcp.NewTool("kiali_namespace_metrics",
				mcp.WithDescription("Get aggregated request rate, error ratio, and latency metrics for all services in a namespace. Use as the namespace-level triage step — call this before kiali_service_metrics to identify which specific services are degraded rather than querying each service individually. Use kiali_mesh_graph first when you need to understand failure propagation paths across namespaces before drilling into a single namespace."),
				mcp.WithString("namespace", mcp.Description("Namespace to get metrics for"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiNamespaceMetrics},
		{Tool: WithMeta(
			mcp.NewTool("kiali_service_metrics",
				mcp.WithDescription("Get granular traffic metrics for a specific service: request volumes, error rates, latency distributions, and traffic composition. Call after kiali_namespace_metrics has identified which service is degraded — this provides the per-service detail needed to correlate with recent deployments or config changes. Use kiali_clusters_services first when the exact service name is unknown."),
				mcp.WithString("namespace", mcp.Description("Namespace of the service"), mcp.Required()),
				mcp.WithString("service", mcp.Description("Service name"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiServiceMetrics},
		{Tool: WithMeta(
			mcp.NewTool("kiali_mesh_graph",
				mcp.WithDescription("Retrieve the full service mesh topology graph with real-time health and mTLS status for all workloads. Maps inter-service communication paths, traffic flow bottlenecks, and failure propagation. Call this first at the start of a mesh-wide incident to understand blast radius and identify which upstream services are affected — then drill into specific namespaces with kiali_namespace_metrics and kiali_namespace_validations. Returns a potentially large graph; use for topology orientation, not per-service metrics."),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiMeshGraph},
	}
}

// kialiHealthCheck handles the kiali_health_check tool request
func (s *Server) kialiHealthCheck(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_health_check failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: kiali_health_check - checking Kiali health status - got called by session id: %s", sessionID)

	ret, err := k.KialiHealthCheck(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_health_check failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to check Kiali health: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_health_check completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiNamespaceValidations handles the kiali_namespace_validations tool request
func (s *Server) kialiNamespaceValidations(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_namespace_validations failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: kiali_namespace_validations - getting validations for namespace: %s - got called by session id: %s", namespace, sessionID)

	ret, err := k.GetNamespaceValidations(ctx, namespace)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_namespace_validations failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get namespace validations: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_namespace_validations completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiNamespaceTLS handles the kiali_namespace_tls tool request
func (s *Server) kialiNamespaceTLS(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_namespace_tls failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: kiali_namespace_tls - getting TLS status for namespace: %s - got called by session id: %s", namespace, sessionID)

	ret, err := k.GetNamespaceTLS(ctx, namespace)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_namespace_tls failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get namespace TLS status: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_namespace_tls completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiClustersServices handles the kiali_clusters_services tool request
func (s *Server) kialiClustersServices(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_clusters_services failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: kiali_clusters_services - getting clusters services for namespace: %s - got called by session id: %s", namespace, sessionID)

	ret, err := k.GetClustersServices(ctx, namespace)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_clusters_services failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get clusters services: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_clusters_services completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiIstioConfig handles the kiali_istio_config tool request
func (s *Server) kialiIstioConfig(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_istio_config failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)

	klog.V(1).Infof("Tool: kiali_istio_config - getting Istio configuration - got called by session id: %s", sessionID)

	ret, err := k.GetKialiIstioConfig(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_istio_config failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get Istio configuration: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_istio_config completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiTracingInfo handles the kiali_tracing_info tool request
func (s *Server) kialiTracingInfo(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_tracing_info failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)

	klog.V(1).Infof("Tool: kiali_tracing_info - getting tracing information - got called by session id: %s", sessionID)

	ret, err := k.GetKialiTracingInfo(ctx)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_tracing_info failed after %v: %v by session id: %s", duration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get tracing information: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_tracing_info completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiNamespaceMetrics handles the kiali_namespace_metrics tool request
func (s *Server) kialiNamespaceMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_namespace_metrics failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	namespace := ctr.GetString("namespace", "")

	klog.V(1).Infof("Tool: kiali_namespace_metrics - getting metrics for namespace: %s - got called by session id: %s", namespace, sessionID)

	ret, err := k.GetKialiNamespaceMetrics(ctx, namespace)
	requestDuration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_namespace_metrics failed after %v: %v by session id: %s", requestDuration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get namespace metrics: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_namespace_metrics completed successfully in %v by session id: %s", requestDuration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiServiceMetrics handles the kiali_service_metrics tool request
func (s *Server) kialiServiceMetrics(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_service_metrics failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)
	namespace := ctr.GetString("namespace", "")
	service := ctr.GetString("service", "")

	klog.V(1).Infof("Tool: kiali_service_metrics - getting metrics for service: %s in namespace: %s - got called by session id: %s", service, namespace, sessionID)

	ret, err := k.GetKialiServiceMetrics(ctx, namespace, service)
	requestDuration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_service_metrics failed after %v: %v by session id: %s", requestDuration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get service metrics: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_service_metrics completed successfully in %v by session id: %s", requestDuration, sessionID)
	return NewTextResult(ret, nil), nil
}

// kialiMeshGraph handles the kiali_mesh_graph tool request
func (s *Server) kialiMeshGraph(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: kiali_mesh_graph failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	sessionID := getSessionID(ctx)

	klog.V(1).Infof("Tool: kiali_mesh_graph - getting mesh graph - got called by session id: %s", sessionID)

	ret, err := k.GetKialiMeshGraph(ctx)
	requestDuration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: kiali_mesh_graph failed after %v: %v by session id: %s", requestDuration, err, sessionID)
		return NewTextResult("", fmt.Errorf("failed to get mesh graph: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: kiali_mesh_graph completed successfully in %v by session id: %s", requestDuration, sessionID)
	return NewTextResult(ret, nil), nil
}
