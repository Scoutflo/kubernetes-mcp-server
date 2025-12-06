package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// Risk levels for HITL
const (
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
	RiskCritical = "critical"
)

// WithHITLMeta adds HITL metadata to a tool
func WithHITLMeta(tool mcp.Tool, riskLevel, message string) mcp.Tool {
	tool.Meta = mcp.NewMetaFromMap(map[string]any{
		"hitl/required":     true,
		"hitl/riskLevel":    riskLevel,
		"hitl/approvalType": "single",
		"hitl/message":      message,
	})
	return tool
}


