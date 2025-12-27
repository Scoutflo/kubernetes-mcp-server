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
// Automatically transforms keys inside nested `hitl` objects to have `hitl/` prefix for LangGraph compatibility.
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

	// Process incoming metadata: handle flat hitl/* keys and nested hitl objects
	transformedMeta := make(map[string]any)
	hitlMap := make(map[string]any)

	// Check if existing metadata has a nested hitl object
	if existingHitlRaw, ok := existingMeta["hitl"]; ok {
		if existingHitlMap, ok := existingHitlRaw.(map[string]any); ok {
			// Copy existing hitl values (they may already have hitl/ prefix)
			for k, v := range existingHitlMap {
				hitlMap[k] = v
			}
		}
	}

	// Process incoming metadata
	for k, v := range metadata {
		if strings.HasPrefix(k, "hitl/") {
			// Flat hitl/* key at top level - add to nested hitl object with hitl/ prefix
			hitlMap[k] = v
		} else if k == "hitl" {
			// Nested hitl object - process its keys
			if hitlObj, ok := v.(map[string]any); ok {
				for hitlKey, hitlValue := range hitlObj {
					// If key doesn't already have hitl/ prefix, add it
					if !strings.HasPrefix(hitlKey, "hitl/") {
						hitlMap["hitl/"+hitlKey] = hitlValue
					} else {
						hitlMap[hitlKey] = hitlValue
					}
				}
			}
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

	// Transform any existing nested hitl object in merged metadata
	if hitlObj, ok := existingMeta["hitl"].(map[string]any); ok {
		transformedHitl := make(map[string]any)
		for k, v := range hitlObj {
			// If key doesn't already have hitl/ prefix, add it
			if !strings.HasPrefix(k, "hitl/") {
				transformedHitl["hitl/"+k] = v
			} else {
				transformedHitl[k] = v
			}
		}
		existingMeta["hitl"] = transformedHitl
	}

	// Set the merged metadata
	tool.Meta = mcp.NewMetaFromMap(existingMeta)
	return tool
}
