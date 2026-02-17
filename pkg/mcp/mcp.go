package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/health"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/kubernetes"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/version"
	"k8s.io/klog/v2"
)

const (
	// HealthPort is the port used for health checks
	HealthPort = 8082
)

type K8sCredentials struct {
	K8sURL   string `json:"k8surl"`
	K8sToken string `json:"k8stoken"`
}

type Server struct {
	server      *server.MCPServer
	healthCheck *health.HealthChecker
}

func NewSever() (*Server, error) {
	hooks := &server.Hooks{
		OnRegisterSession: []server.OnRegisterSessionHookFunc{
			func(ctx context.Context, session server.ClientSession) {
				klog.V(0).Infof("New MCP session registered: %s", session.SessionID())
			},
		},
		OnUnregisterSession: []server.OnUnregisterSessionHookFunc{
			func(ctx context.Context, session server.ClientSession) {
				klog.V(0).Infof("MCP session unregistered: %s", session.SessionID())
			},
		},
	}

	s := &Server{
		server: server.NewMCPServer(
			version.BinaryName,
			version.Version,
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
			server.WithHooks(hooks),
		),
		healthCheck: health.NewHealthChecker(),
	}

	klog.V(0).Infof("Initializing MCP server %s version %s", version.BinaryName, version.Version)

	if err := s.initializeTools(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) initializeTools() error {
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
		s.initKiali(),
	)

	s.server.SetTools(tools...)
	klog.V(0).Infof("Registered %d MCP tools", len(tools))

	return nil
}

func (s *Server) ServeSse(baseUrl string) *server.SSEServer {
	options := make([]server.SSEOption, 0)
	if baseUrl != "" {
		options = append(options, server.WithBaseURL(baseUrl))
	}

	options = append(options, server.WithKeepAlive(true))
	options = append(options, server.WithKeepAliveInterval(30*time.Minute))

	sseServer := server.NewSSEServer(s.server, options...)

	go s.startHealthServer()

	go func() {
		s.healthCheck.SetReady(true)
	}()

	return sseServer
}

func (s *Server) startHealthServer() {
	mux := http.NewServeMux()
	health.AttachHealthEndpoints(mux, s.healthCheck)

	healthServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", HealthPort),
		Handler: mux,
	}

	if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		klog.Errorf("Health server error: %v", err)
	}
}

func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		errMsg := err.Error()
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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}

// getSessionID extracts the session ID from context for logging purposes
func getSessionID(ctx context.Context) string {
	if session := server.ClientSessionFromContext(ctx); session != nil {
		return session.SessionID()
	}
	return "unknown"
}

// getKubernetesClient creates or returns a Kubernetes client based on the tool request parameters
// If k8surl and k8stoken are provided in the request, it creates a new client with those credentials
// Otherwise, it tries to use environment variables as fallback
func (s *Server) getKubernetesClient(ctr mcp.CallToolRequest) (*kubernetes.Kubernetes, error) {

	meta, err := ctr.Params.Meta.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta: %v", err)
	}

	klog.V(0).Infof("ctr.Params.Meta: %v", ctr.Params.Meta)

	k8sCredentials := K8sCredentials{}
	if err := json.Unmarshal(meta, &k8sCredentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %v", err)
	}

	if k8sCredentials.K8sURL != "" && k8sCredentials.K8sToken != "" {
		// Create client with provided credentials
		return kubernetes.NewKubernetesWithCredentials(k8sCredentials.K8sURL, k8sCredentials.K8sToken)
	}
}
