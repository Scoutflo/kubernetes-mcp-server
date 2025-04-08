package mcp

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestWatchKubeConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-linux platforms")
	}
	testCase(t, func(c *mcpContext) {
		// Given
		withTimeout, cancel := context.WithTimeout(c.ctx, 5*time.Second)
		defer cancel()
		var notification *mcp.JSONRPCNotification
		c.mcpClient.OnNotification(func(n mcp.JSONRPCNotification) {
			notification = &n
		})
		// When
		f, _ := os.OpenFile(filepath.Join(c.tempDir, "config"), os.O_APPEND|os.O_WRONLY, 0644)
		_, _ = f.WriteString("\n")
		for {
			if notification != nil {
				break
			}
			select {
			case <-withTimeout.Done():
				break
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
		// Then
		t.Run("WatchKubeConfig notifies tools change", func(t *testing.T) {
			if notification == nil {
				t.Fatalf("WatchKubeConfig did not notify")
			}
			if notification.Method != "notifications/tools/list_changed" {
				t.Fatalf("WatchKubeConfig did not notify tools change, got %s", notification.Method)
			}
		})
	})
}

func TestTools(t *testing.T) {
	expectedNames := []string{
		"configuration_view",
		"events_list",
		"namespaces_list",
		"pods_list",
		"pods_list_in_namespace",
		"pods_get",
		"pods_delete",
		"pods_log",
		"pods_run",
		"resources_list",
		"resources_get",
		"resources_create_or_update",
		"resources_delete",
	}
	testCase(t, func(c *mcpContext) {
		tools, err := c.mcpClient.ListTools(c.ctx, mcp.ListToolsRequest{})
		t.Run("ListTools returns tools", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call ListTools failed %v", err)
				return
			}
		})
		nameSet := make(map[string]bool)
		for _, tool := range tools.Tools {
			nameSet[tool.Name] = true
		}
		for _, name := range expectedNames {
			t.Run("ListTools has "+name+" tool", func(t *testing.T) {
				if nameSet[name] != true {
					t.Fatalf("tool %s not found", name)
					return
				}
			})
		}
	})
}
