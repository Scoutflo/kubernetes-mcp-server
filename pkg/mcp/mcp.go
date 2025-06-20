package mcp

import (
	"github.com/manusa/kubernetes-mcp-server/pkg/kubernetes"
	"github.com/manusa/kubernetes-mcp-server/pkg/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"slices"
)

type Server struct {
	server *server.MCPServer
	k      *kubernetes.Kubernetes
}

func NewSever() (*Server, error) {
	s := &Server{
		server: server.NewMCPServer(
			version.BinaryName,
			version.Version,
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
		),
	}
	if err := s.reloadKubernetesClient(); err != nil {
		return nil, err
	}
	s.k.WatchKubeConfig(s.reloadKubernetesClient)
	return s, nil
}

func (s *Server) reloadKubernetesClient() error {
	k, err := kubernetes.NewKubernetes()
	if err != nil {
		return err
	}
	s.k = k
	s.server.SetTools(slices.Concat(
		s.initConfiguration(),
		s.initEvents(),
		s.initNamespaces(),
		s.initPods(),
		s.initResources(),
	)...)
	return nil
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}

func (s *Server) ServeSse(baseUrl string) *server.SSEServer {
	options := make([]server.SSEOption, 0)
	if baseUrl != "" {
		options = append(options, server.WithBaseURL(baseUrl))
		options = append(options, server.WithKeepAlive(true))
	}
	return server.NewSSEServer(s.server, options...).
}

func (s *Server) Close() {
	if s.k != nil {
		s.k.Close()
	}
}

func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: err.Error(),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}
