package mcp

import (
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// KubernetesSession implements the ClientSession interface and optionally
// the SessionWithTools interface for the Kubernetes MCP server.
type KubernetesSession struct {
	id            string
	notifChannel  chan mcp.JSONRPCNotification
	isInitialized bool
	sessionTools  map[string]server.ServerTool
	toolsMutex    sync.RWMutex
}

// NewKubernetesSession creates a new session with the given ID
func NewKubernetesSession(id string) *KubernetesSession {
	return &KubernetesSession{
		id:           id,
		notifChannel: make(chan mcp.JSONRPCNotification, 10),
		sessionTools: make(map[string]server.ServerTool),
	}
}

// SessionID returns the unique identifier for this session
func (s *KubernetesSession) SessionID() string {
	return s.id
}

// NotificationChannel returns the channel for sending notifications to this session
func (s *KubernetesSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return s.notifChannel
}

// Initialize marks the session as fully initialized
func (s *KubernetesSession) Initialize() {
	s.isInitialized = true
}

// Initialized returns whether the session is initialized
func (s *KubernetesSession) Initialized() bool {
	return s.isInitialized
}

// GetSessionTools returns the tools specific to this session
func (s *KubernetesSession) GetSessionTools() map[string]server.ServerTool {
	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()
	return s.sessionTools
}

// SetSessionTools sets the tools specific to this session
func (s *KubernetesSession) SetSessionTools(tools map[string]server.ServerTool) {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()
	s.sessionTools = tools
}
