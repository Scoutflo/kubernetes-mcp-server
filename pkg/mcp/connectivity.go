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
		{Tool: mcp.NewTool("check_service_connectivity",
			mcp.WithDescription("Check connectivity to a Kubernetes service"),
			mcp.WithString("service_name",
				mcp.Description("Fully qualified service name with port number (e.g. my-service.my-namespace.svc.cluster.local:80)"),
				mcp.Required(),
			),
		), Handler: s.checkServiceConnectivity},
		{Tool: mcp.NewTool("check_ingress_connectivity",
			mcp.WithDescription("Check connectivity to a Kubernetes ingress host"),
			mcp.WithString("ingress_host",
				mcp.Description("Ingress host to check connectivity to (e.g. example.com or https://example.com)"),
				mcp.Required(),
			),
		), Handler: s.checkIngressConnectivity},
	}
}

func (s *Server) checkServiceConnectivity(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	serviceName, err := ctr.RequireString("service_name")
	if err != nil {
		klog.Errorf("Tool call: check_service_connectivity failed after %v: missing or invalid service_name", time.Since(start))
		return NewTextResult("", errors.New("failed to check service connectivity, missing or invalid service_name")), nil
	}

	result, err := s.k.CheckServiceConnectivity(ctx, serviceName)
	if err != nil {
		return NewTextResult("", fmt.Errorf("connectivity check failed: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: check_service_connectivity completed successfully in %v", time.Since(start))
	return NewTextResult(result, nil), nil
}

func (s *Server) checkIngressConnectivity(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	ingressHost, err := ctr.RequireString("ingress_host")
	if err != nil {
		klog.Errorf("Tool call: check_ingress_connectivity failed after %v: missing or invalid ingress_host", time.Since(start))
		return NewTextResult("", errors.New("failed to check ingress connectivity, missing or invalid ingress_host")), nil
	}

	result, err := s.k.CheckIngressConnectivity(ctx, ingressHost)
	if err != nil {
		return NewTextResult("", fmt.Errorf("ingress connectivity check failed: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: check_ingress_connectivity completed successfully in %v", time.Since(start))
	return NewTextResult(result, nil), nil
}
