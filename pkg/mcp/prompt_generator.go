package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	// Extract the description parameter
	descriptionArg := ctr.Params.Arguments["description"]
	if descriptionArg == nil {
		return NewTextResult("", errors.New("missing required parameter: description")), nil
	}
	description := descriptionArg.(string)

	// Call the Kubernetes GeneratePrompt function
	prompt, err := s.k.GeneratePrompt(description)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to generate prompt: %v", err)), nil
	}

	// Return the formatted result
	return NewTextResult(prompt, nil), nil
}
