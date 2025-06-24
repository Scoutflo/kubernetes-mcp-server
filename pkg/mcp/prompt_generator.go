package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/klog/v2"
)

// initPromptGenerator initializes the prompt generator and returns the server tools
func (s *Server) initPromptGenerator() []server.ServerTool {

	// Define and return the server tools
	return []server.ServerTool{
		{Tool: mcp.NewTool("prompt_generator",
			mcp.WithDescription("Generate a well-structured prompt for Kubernetes analysis based on a description"),
			mcp.WithString("description",
				mcp.Description("Natural language description of the prompt to generate"),
				mcp.Required(),
			),
		), Handler: s.promptGenerator},
	}
}

func (s *Server) promptGenerator(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()

	// Extract the description parameter
	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		duration := time.Since(start)
		klog.Errorf("Tool call: prompt_generator failed after %v: failed to get arguments", duration)
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	descriptionArg, exists := argsMap["description"]
	if !exists || descriptionArg == nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prompt_generator failed after %v: missing required parameter: description", duration)
		return NewTextResult("", errors.New("missing required parameter: description")), nil
	}
	description := descriptionArg.(string)

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: prompt_generator - description=%s - got called by session id: %s", description, sessionID)

	// Call the Kubernetes GeneratePrompt function
	prompt, err := s.k.GeneratePrompt(description)
	if err != nil {
		duration := time.Since(start)
		klog.Errorf("Tool call: prompt_generator failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to generate prompt: %v", err)), nil
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: prompt_generator completed successfully in %v", duration)
	return NewTextResult(prompt, nil), nil
}
