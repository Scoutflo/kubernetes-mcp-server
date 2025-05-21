package kubernetes

import (
	"fmt"

	"github.com/scoutflo/kubernetes-mcp-server/pkg/llm"
)

// GeneratePrompt generates a structured prompt based on a description
func (k *Kubernetes) GeneratePrompt(description string) (string, error) {
	llmClient, err := llm.NewDefaultClient()
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %v", err)
	}

	response, err := llmClient.Call(llm.PromptGeneratorKnowledgeBase, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate prompt: %v", err)
	}
	return response, nil
}
