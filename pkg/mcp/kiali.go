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
				mcp.WithDescription("Check the health of Kiali service. Verifies if Kiali is accessible and responding properly."),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiHealthCheck},
		{Tool: WithMeta(
			mcp.NewTool("kiali_namespace_validations",
				mcp.WithDescription("Namespace Validation Auditor - Detects Istio configuration errors before they cause outages. Scans VirtualServices, DestinationRules, and policies for semantic/runtime errors like invalid hosts or conflicting rules. Provides actionable diagnostics to prevent traffic disruptions and accelerate troubleshooting."),
				mcp.WithString("namespace", mcp.Description("Namespace to get validations for"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiNamespaceValidations},
		{Tool: WithMeta(
			mcp.NewTool("kiali_namespace_tls",
				mcp.WithDescription("Namespace TLS Compliance Checker - Audits the mTLS encryption status across namespace workloads. Evaluates traffic encryption state (STRICT/PERMISSIVE/DISABLED) by synthesizing PeerAuthentications and DestinationRules. Identifies security gaps for compliance enforcement and decrypts TLS-related failure root causes."),
				mcp.WithString("namespace", mcp.Description("Namespace to get TLS status for"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiNamespaceTLS},
		{Tool: WithMeta(
			mcp.NewTool("kiali_clusters_services",
				mcp.WithDescription("Cluster Service Inventory Tool - Discover all services with Istio sidecar status across clusters. Catalogue services with metadata, labels, and mesh participation status. Enables mesh coverage auditing, orphaned service detection, and dependency mapping."),
				mcp.WithString("namespace", mcp.Description("Filter by specific namespace (optional)")),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiClustersServices},
		{Tool: WithMeta(
			mcp.NewTool("kiali_istio_config",
				mcp.WithDescription("Istio Configuration Inventory Tool - Audits all Istio config objects for drift and compliance. Lists VirtualServices, DestinationRules, Gateways, and ServiceEntries with normalized specs. Detects configuration drift, undocumented changes, and policy violations."),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiIstioConfig},
		{Tool: WithMeta(
			mcp.NewTool("kiali_tracing_info",
				mcp.WithDescription("Tracing Integration Status Tool - Verifies distributed tracing backend connectivity and health. Reports tracing system status (Jaeger/Tempo), endpoint availability, and configuration errors. Ensures trace data exists for incident investigations."),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiTracingInfo},
		{Tool: WithMeta(
			mcp.NewTool("kiali_namespace_metrics",
				mcp.WithDescription("Namespace Metrics Aggregator - Provides a namespace-level health and performance overview. Aggregates workload request rates, error ratios, and latency across all services in a namespace. Detects anomalous patterns and degradation hotspots."),
				mcp.WithString("namespace", mcp.Description("Namespace to get metrics for"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiNamespaceMetrics},
		{Tool: WithMeta(
			mcp.NewTool("kiali_service_metrics",
				mcp.WithDescription("Service Metrics Explorer - Delivers granular service-level traffic and error analytics. Reveals request volumes, failure rates, latency distributions, and traffic composition for individual services. Enables deep dives during outages and performance optimization."),
				mcp.WithString("namespace", mcp.Description("Namespace of the service"), mcp.Required()),
				mcp.WithString("service", mcp.Description("Service name"), mcp.Required()),
			),
			map[string]any{"provider": ProviderKiali},
		), Handler: s.kialiServiceMetrics},
		{Tool: WithMeta(
			mcp.NewTool("kiali_mesh_graph",
				mcp.WithDescription("Mesh Topology Mapper - Visualizes service dependencies with real-time health and security context. Maps workload communications, traffic flow bottlenecks, mTLS status, and failure propagation paths. Essential for blast radius analysis, root cause identification, and security exposure assessment."),
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
