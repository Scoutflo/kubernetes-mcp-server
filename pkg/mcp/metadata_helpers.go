package mcp

import (
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Risk levels for HITL
const (
	RiskLow      = "low"
	RiskMedium   = "medium"
	RiskHigh     = "high"
	RiskCritical = "critical"
)

// Provider constants
const (
	ProviderKubernetes   = "kubernetes"
	ProviderPrometheus   = "prometheus"
	ProviderGrafana      = "grafana"
	ProviderArgoCD       = "argocd"
	ProviderArgoRollouts = "argorollouts"
	ProviderHelm         = "helm"
	ProviderIstio        = "istio"
	ProviderKiali        = "kiali"
)

// HITLConfig represents HITL metadata as a map with type-safe accessors
type HITLConfig map[string]interface{}

// Bool safely retrieves a boolean value from HITLConfig
func (h HITLConfig) Bool(key string) bool {
	if val, ok := h[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// String safely retrieves a string value from HITLConfig
func (h HITLConfig) String(key string) string {
	if val, ok := h[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// Has checks if a key exists in HITLConfig
func (h HITLConfig) Has(key string) bool {
	_, ok := h[key]
	return ok
}

// WithMeta adds or merges metadata to a tool. If the tool already has metadata,
// it will be merged with the new metadata (new values override existing ones for same keys).
// Automatically transforms flat keys with `hitl/` prefix to nested `hitl` object structure.
func WithMeta(tool mcp.Tool, metadata map[string]any) mcp.Tool {
	// Start with existing metadata if it exists
	existingMeta := make(map[string]any)
	if tool.Meta != nil {
		// Extract existing metadata by marshaling and unmarshaling
		metaJSON, err := tool.Meta.MarshalJSON()
		if err == nil {
			json.Unmarshal(metaJSON, &existingMeta)
		}
	}

	// Transform incoming metadata: convert flat `hitl/*` keys to nested `hitl` object
	transformedMeta := make(map[string]any)
	hitlMap := make(map[string]any)

	// Check if metadata already has a nested hitl object
	var existingHitl map[string]any
	if existingHitlRaw, ok := existingMeta["hitl"]; ok {
		if existingHitlMap, ok := existingHitlRaw.(map[string]any); ok {
			existingHitl = existingHitlMap
			// Copy existing hitl values
			for k, v := range existingHitl {
				hitlMap[k] = v
			}
		}
	}

	// Process incoming metadata
	for k, v := range metadata {
		if strings.HasPrefix(k, "hitl/") {
			// Extract key without "hitl/" prefix
			hitlKey := strings.TrimPrefix(k, "hitl/")
			hitlMap[hitlKey] = v
		} else {
			// Keep non-hitl keys as-is
			transformedMeta[k] = v
		}
	}

	// Add hitl object if it has any values
	if len(hitlMap) > 0 {
		transformedMeta["hitl"] = hitlMap
	}

	// Merge transformed metadata with existing metadata
	for k, v := range transformedMeta {
		existingMeta[k] = v
	}

	// Set the merged metadata
	tool.Meta = mcp.NewMetaFromMap(existingMeta)
	return tool
}
