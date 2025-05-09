package mcp

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventsList(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		c.withEnvTest()
		toolResult, err := c.callTool("events_list", map[string]interface{}{})
		t.Run("events_list with no events returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if toolResult.Content[0].(mcp.TextContent).Text != "No events found" {
				t.Fatalf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
		client := c.newKubernetesClient()
		for _, ns := range []string{"default", "ns-1"} {
			_, _ = client.CoreV1().Events(ns).Create(c.ctx, &v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "an-event-in-" + ns,
				},
				InvolvedObject: v1.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "a-pod",
					Namespace:  ns,
				},
				Type:    "Normal",
				Message: "The event message",
			}, metav1.CreateOptions{})
		}
		toolResult, err = c.callTool("events_list", map[string]interface{}{})
		t.Run("events_list with events returns all OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if toolResult.Content[0].(mcp.TextContent).Text != "The following events (YAML format) were found:\n"+
				"- InvolvedObject:\n"+
				"    Kind: Pod\n"+
				"    Name: a-pod\n"+
				"    apiVersion: v1\n"+
				"  Message: The event message\n"+
				"  Namespace: default\n"+
				"  Reason: \"\"\n"+
				"  Timestamp: 0001-01-01 00:00:00 +0000 UTC\n"+
				"  Type: Normal\n"+
				"- InvolvedObject:\n"+
				"    Kind: Pod\n"+
				"    Name: a-pod\n"+
				"    apiVersion: v1\n"+
				"  Message: The event message\n"+
				"  Namespace: ns-1\n"+
				"  Reason: \"\"\n"+
				"  Timestamp: 0001-01-01 00:00:00 +0000 UTC\n"+
				"  Type: Normal\n" {
				t.Fatalf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
		toolResult, err = c.callTool("events_list", map[string]interface{}{
			"namespace": "ns-1",
		})
		t.Run("events_list in namespace with events returns from namespace OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if toolResult.Content[0].(mcp.TextContent).Text != "The following events (YAML format) were found:\n"+
				"- InvolvedObject:\n"+
				"    Kind: Pod\n"+
				"    Name: a-pod\n"+
				"    apiVersion: v1\n"+
				"  Message: The event message\n"+
				"  Namespace: ns-1\n"+
				"  Reason: \"\"\n"+
				"  Timestamp: 0001-01-01 00:00:00 +0000 UTC\n"+
				"  Type: Normal\n" {
				t.Fatalf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})

		// Test field selectors
		toolResult, err = c.callTool("events_list", map[string]interface{}{
			"involved_object_name": "a-pod",
			"involved_object_kind": "Pod",
		})
		t.Run("events_list with field selectors returns filtered events OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if !containsEventWithNamespace(toolResult.Content[0].(mcp.TextContent).Text, "default") ||
				!containsEventWithNamespace(toolResult.Content[0].(mcp.TextContent).Text, "ns-1") {
				t.Fatalf("expected to find events from both namespaces, got: %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
	})
}

// Helper function to check if the result contains an event from a specific namespace
func containsEventWithNamespace(result string, namespace string) bool {
	return strings.Contains(result, "Namespace: "+namespace)
}
