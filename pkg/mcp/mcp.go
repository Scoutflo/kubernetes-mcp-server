package mcp

import (
	"slices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/version"
)

// ServerMode represents the mode the server is running in
type ServerMode int

const (
	// UnknownMode is the default mode before the server starts
	UnknownMode ServerMode = iota
	// StdioMode indicates the server is running in STDIO mode
	StdioMode
	// SseMode indicates the server is running in SSE mode
	SseMode
)

type Server struct {
	server *server.MCPServer
	k      *kubernetes.Kubernetes
	mode   ServerMode
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
		mode: UnknownMode,
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
		s.initRollouts(),
		s.initNamespaces(),
		s.initPods(),
		s.initResources(),
		s.initPortForward(),
		s.initNodes(),
		s.initMetricsServer(),
		s.initPrometheus(),
		s.initLabels(),
		s.initConnectivity(),
		s.initHelm(),
		s.initArgoCD(),
		s.initArgoRollouts(),
	)...)
	return nil
}

// GetServerMode returns the current server mode
func (s *Server) GetServerMode() ServerMode {
	return s.mode
}

// IsStdioMode returns true if the server is running in STDIO mode
func (s *Server) IsStdioMode() bool {
	return s.mode == StdioMode
}

func (s *Server) ServeStdio() error {
	s.mode = StdioMode
	return server.ServeStdio(s.server)
}

func (s *Server) ServeSse(baseUrl string) *server.SSEServer {
	s.mode = SseMode
	options := make([]server.SSEOption, 0)
	if baseUrl != "" {
		options = append(options, server.WithBaseURL(baseUrl))
	}
	return server.NewSSEServer(s.server, options...)
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
