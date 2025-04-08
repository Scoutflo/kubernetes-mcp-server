package mcp

import (
	"slices"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestNamespacesList(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		c.withEnvTest()
		toolResult, err := c.callTool("namespaces_list", map[string]interface{}{})
		t.Run("namespaces_list returns namespace list", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
		})
		var decoded []unstructured.Unstructured
		err = yaml.Unmarshal([]byte(toolResult.Content[0].(mcp.TextContent).Text), &decoded)
		t.Run("namespaces_list has yaml content", func(t *testing.T) {
			if err != nil {
				t.Fatalf("invalid tool result content %v", err)
			}
		})
		t.Run("namespaces_list returns at least 3 items", func(t *testing.T) {
			if len(decoded) < 3 {
				t.Errorf("invalid namespace count, expected at least 3, got %v", len(decoded))
			}
			for _, expectedNamespace := range []string{"default", "ns-1", "ns-2"} {
				idx := slices.IndexFunc(decoded, func(ns unstructured.Unstructured) bool {
					return ns.GetName() == expectedNamespace
				})
				if idx == -1 {
					t.Errorf("namespace %s not found in the list", expectedNamespace)
				}
			}
		})
	})
}
