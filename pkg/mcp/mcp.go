package mcp

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/health"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/version"
	"k8s.io/klog/v2"
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

	// HealthPort is the port used for health checks
	HealthPort = 8082
)

type Server struct {
	server       *server.MCPServer
	k            *kubernetes.Kubernetes
	mode         ServerMode
	healthCheck  *health.HealthChecker
	healthServer *http.Server
	sessions     sync.Map
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
		mode:        UnknownMode,
		healthCheck: health.NewHealthChecker(),
	}

	klog.V(0).Infof("Initializing MCP server %s version %s", version.BinaryName, version.Version)

	if err := s.initializeKubernetesClient(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) initializeKubernetesClient() error {
	klog.V(0).Infof("Initializing Kubernetes client...")

	k, err := kubernetes.NewKubernetes()
	if err != nil {
		klog.Errorf("Failed to initialize Kubernetes client: %v", err)
		return err
	}
	s.k = k

	klog.V(0).Infof("Registering MCP tools...")

	tools := slices.Concat(
		s.initConfiguration(),
		s.initEvents(),
		s.initRollouts(),
		s.initNamespaces(),
		s.initPods(),
		s.initResources(),
		s.initNodes(),
		s.initMetricsServer(),
		s.initPrometheus(),
		s.initLabels(),
		s.initConnectivity(),
		s.initHelm(),
		s.initArgoCD(),
		s.initArgoRollouts(),
		s.initConverters(),
		s.initPromptGenerator(),
		s.initGrafana(),
		s.initIstio(),
	)

	s.server.SetTools(tools...)
	klog.V(0).Infof("Registered %d MCP tools", len(tools))

	// Initialize MCP resources for Kubernetes documentation
	s.initDocumentationResources()
	klog.V(0).Infof("Kubernetes MCP server initialization complete")

	return nil
}

// reloadKubernetesClient for legacy kubeconfig watching (only used in non-HTTP mode)
func (s *Server) reloadKubernetesClient() error {
	return s.initializeKubernetesClient()
}

// GetServerMode returns the current server mode
func (s *Server) GetServerMode() ServerMode {
	return s.mode
}

// IsStdioMode returns true if the server is running in STDIO mode
func (s *Server) IsStdioMode() bool {
	return s.mode == StdioMode
}

// CreateSession creates a new session with a unique ID
func (s *Server) CreateSession() (*KubernetesSession, error) {
	sessionID := uuid.New().String()
	session := NewKubernetesSession(sessionID)

	// Register the session with the MCP server
	if err := s.server.RegisterSession(context.Background(), session); err != nil {
		return nil, fmt.Errorf("failed to register session: %w", err)
	}

	// Store in our local sessions map for tracking
	s.sessions.Store(sessionID, session)

	return session, nil
}

// GetSession retrieves a session by ID
func (s *Server) GetSession(sessionID string) (*KubernetesSession, bool) {
	if sessionVal, ok := s.sessions.Load(sessionID); ok {
		if session, ok := sessionVal.(*KubernetesSession); ok {
			return session, true
		}
	}
	return nil, false
}

// DestroySession removes a session
func (s *Server) DestroySession(sessionID string) {
	s.sessions.Delete(sessionID)
	s.server.UnregisterSession(context.Background(), sessionID)
}

// AddSessionTool adds a tool specific to a session
func (s *Server) AddSessionTool(sessionID string, tool mcp.Tool, handler server.ToolHandlerFunc) error {
	return s.server.AddSessionTool(sessionID, tool, handler)
}

// DeleteSessionTools removes tools from a session
func (s *Server) DeleteSessionTools(sessionID string, toolNames ...string) error {
	return s.server.DeleteSessionTools(sessionID, toolNames...)
}

func (s *Server) ServeStdio() error {
	s.mode = StdioMode

	// For STDIO mode, create a single session
	session, err := s.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session for STDIO mode: %w", err)
	}

	// Set the context with the session but don't use it directly
	// The session is registered with the server and will be used for notifications
	_ = s.server.WithContext(context.Background(), session)

	// Use the standard ServeStdio function
	return server.ServeStdio(s.server)
}

func (s *Server) ServeSse(baseUrl string) *server.SSEServer {
	s.mode = SseMode
	options := make([]server.SSEOption, 0)
	if baseUrl != "" {
		options = append(options, server.WithBaseURL(baseUrl))
	}

	// Configure keep-alive settings for long-lived SSE connections
	options = append(options, server.WithKeepAlive(true))
	options = append(options, server.WithKeepAliveInterval(30*time.Minute))

	// Create the SSE server with the configured options
	sseServer := server.NewSSEServer(s.server, options...)

	// Start the health check server on the health port
	go s.startHealthServer()

	// Mark as ready after a short delay to allow server to initialize
	go func() {
		// Set ready status after server starts
		s.healthCheck.SetReady(true)
	}()

	return sseServer
}

// contextKeySessionID is a key for storing session ID in context
type contextKeySessionID struct{}

// startHealthServer starts a separate HTTP server for health checks
func (s *Server) startHealthServer() {
	mux := http.NewServeMux()
	health.AttachHealthEndpoints(mux, s.healthCheck)

	// Create health server on the health port
	s.healthServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", HealthPort),
		Handler: mux,
	}

	// Start server
	if err := s.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// Log error but don't crash - use klog to respect output configuration
		klog.Errorf("Health server error: %v", err)
	}
}

func (s *Server) Close() {
	// Clean up all sessions
	s.sessions.Range(func(key, value interface{}) bool {
		if sessionID, ok := key.(string); ok {
			s.DestroySession(sessionID)
		}
		return true
	})

	// Close health server if it exists
	if s.healthServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.healthServer.Shutdown(ctx); err != nil {
			// Log error but don't crash - use klog to respect output configuration
			klog.Errorf("Health server shutdown error: %v", err)
		}
	}
}

func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		errMsg := err.Error()
		// If the error doesn't already have an ERROR prefix, add one to make it more obvious
		if !strings.HasPrefix(errMsg, "ERROR:") {
			errMsg = "ERROR: " + errMsg
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: errMsg,
				},
			},
		}
	}

	// Return the full content without any truncation
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}

// Tool call logging utilities
func logToolStart(toolName string, params ...interface{}) time.Time {
	start := time.Now()
	if len(params) > 0 {
		klog.V(0).Infof("Tool call: %s - %v", toolName, params)
	} else {
		klog.V(0).Infof("Tool call: %s", toolName)
	}
	return start
}

func logToolSuccess(toolName string, start time.Time) {
	duration := time.Since(start)
	klog.V(0).Infof("Tool call: %s completed successfully in %v", toolName, duration)
}

func logToolError(toolName string, start time.Time, err error) {
	duration := time.Since(start)
	klog.Errorf("Tool call: %s failed after %v: %v", toolName, duration, err)
}
