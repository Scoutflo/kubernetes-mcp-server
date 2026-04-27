package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

func (s *Server) initConnectivity() []server.ServerTool {
	return []server.ServerTool{
		{Tool: WithMeta(
			mcp.NewTool("check_service_connectivity",
				mcp.WithDescription("Test TCP/HTTP connectivity to a Kubernetes service endpoint from within the cluster. Returns reachability status and response details. Use to confirm whether a service is actually reachable after verifying the Service resource exists with resources_get. Requires the fully qualified FQDN with port (e.g., my-service.my-namespace.svc.cluster.local:80) — do not use short names alone. Use after ruling out DNS and service-existence issues to isolate network-policy or sidecar-proxy failures."),
				mcp.WithString("service_name",
					mcp.Description("Fully qualified service name with port number (e.g. my-service.my-namespace.svc.cluster.local:80)"),
					mcp.Required(),
				),
			),
			map[string]any{"provider": ProviderKubernetes},
		), Handler: s.checkServiceConnectivity},
		{Tool: WithMeta(
			mcp.NewTool("check_ingress_connectivity",
				mcp.WithDescription("Test HTTP/HTTPS reachability of a Kubernetes ingress host from outside the cluster. Returns connection status and response details including HTTP status code. Use after confirming the Ingress resource exists and has a valid host rule via resources_get — this tool validates that traffic actually reaches the backend, not just that the config is present. Use to distinguish DNS resolution failures, TLS certificate errors, and 502/503 backend errors. Accepts host-only (example.com) or scheme-prefixed (https://example.com) format."),
				mcp.WithString("ingress_host",
					mcp.Description("Ingress host to check connectivity to (e.g. example.com or https://example.com)"),
					mcp.Required(),
				),
			),
			map[string]any{"provider": ProviderKubernetes},
		), Handler: s.checkIngressConnectivity},
	}
}

func (s *Server) checkServiceConnectivity(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: check_service_connectivity failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	klog.V(1).Infof("Tool: check_service_connectivity - got called by session id: %s", sessionID)
	serviceName, err := ctr.RequireString("service_name")
	if err != nil {
		klog.Errorf("Tool call: check_service_connectivity failed after %v: missing or invalid service_name by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to check service connectivity, missing or invalid service_name")), nil
	}

	result, err := k.CheckServiceConnectivity(ctx, serviceName)
	if err != nil {
		return NewTextResult("", fmt.Errorf("connectivity check failed: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: check_service_connectivity completed successfully in %v by session id: %s", time.Since(start), sessionID)
	return NewTextResult(result, nil), nil
}

func (s *Server) checkIngressConnectivity(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	sessionID := getSessionID(ctx)
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: check_ingress_connectivity failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	klog.V(1).Infof("Tool: check_ingress_connectivity - got called by session id: %s", sessionID)
	ingressHost, err := ctr.RequireString("ingress_host")
	if err != nil {
		klog.Errorf("Tool call: check_ingress_connectivity failed after %v: missing or invalid ingress_host by session id: %s", time.Since(start), sessionID)
		return NewTextResult("", errors.New("failed to check ingress connectivity, missing or invalid ingress_host")), nil
	}

	result, err := k.CheckIngressConnectivity(ctx, ingressHost)
	if err != nil {
		return NewTextResult("", fmt.Errorf("ingress connectivity check failed: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: check_ingress_connectivity completed successfully in %v by session id: %s", time.Since(start), sessionID)
	return NewTextResult(result, nil), nil
}
