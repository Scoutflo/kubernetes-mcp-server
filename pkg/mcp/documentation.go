package mcp

import (
	"context"

	kubernetesdocumentation "github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes-documentation"
)

// initDocumentationResources initializes all Kubernetes documentation resources
// Following the same structural pattern as tools
func (s *Server) initDocumentationResources() {
	// Initialize different categories of documentation
	s.initPodDocumentation()
	s.initWorkloadsDocumentation()
	s.initConfigurationDocumentation()
	s.initStorageDocumentation()
	s.initKubectlDocumentation()
}

// initPodDocumentation initializes Pod-related documentation resources
func (s *Server) initPodDocumentation() {
	kubernetesdocumentation.InitPodDocumentation(s.server)
}

// initWorkloadsDocumentation initializes workloads documentation resources
func (s *Server) initWorkloadsDocumentation() {
	kubernetesdocumentation.InitWorkloadsDocumentation(s.server)
}

// initConfigurationDocumentation initializes configuration documentation resources
func (s *Server) initConfigurationDocumentation() {
	kubernetesdocumentation.InitConfigMapDocumentation(s.server)
}

// initStorageDocumentation initializes storage documentation resources
func (s *Server) initStorageDocumentation() {
	kubernetesdocumentation.InitStorageDocumentation(s.server)
}

// initKubectlDocumentation initializes kubectl command reference
func (s *Server) initKubectlDocumentation() {
	kubernetesdocumentation.InitKubectlDocumentation(s.server)
}

// initDynamicDocumentationResources placeholder for any future dynamic resources
func (s *Server) initDynamicDocumentationResources(ctx context.Context) error {
	// This function is called from mcp.go and can be used to add any
	// dynamic resources that require context or runtime information
	return nil
}
