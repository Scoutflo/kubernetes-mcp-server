package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Config represents the configuration for the LLM client
type Config struct {
	APIKey           string
	Endpoint         string
	DeploymentName   string
	APIVersion       string
	MaxTokens        int
	Temperature      float64
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to the OpenAI Chat API
type ChatRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
}

// ChatResponse represents a response from the OpenAI Chat API
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Client is a client for the LLM API
type Client struct {
	config Config
	client *http.Client
}

// NewClient creates a new LLM client with the given configuration
func NewClient(config Config) *Client {
	// Set default values if not provided
	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.TopP == 0 {
		config.TopP = 1.0
	}

	return &Client{
		config: config,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewDefaultClient creates a new LLM client with configuration from environment variables
func NewDefaultClient() (*Client, error) {
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	deploymentName := os.Getenv("AZURE_OPENAI_DEPLOYMENT_NAME")
	apiVersion := os.Getenv("OPENAI_API_VERSION")
	maxTokensStr := os.Getenv("AZURE_OPENAI_MAX_TOKENS")
	temperatureStr := os.Getenv("AZURE_OPENAI_TEMPERATURE")
	topPStr := os.Getenv("AZURE_OPENAI_TOP_P")

	if apiKey == "" || endpoint == "" || deploymentName == "" || apiVersion == "" || maxTokensStr == "" || temperatureStr == "" || topPStr == "" {
		return nil, errors.New("missing required environment variables for Azure OpenAI")
	}

	// Convert string values to proper types
	maxTokens, err := strconv.Atoi(maxTokensStr)
	if err != nil {
		return nil, fmt.Errorf("invalid value for AZURE_OPENAI_MAX_TOKENS: %v", err)
	}

	temperature, err := strconv.ParseFloat(temperatureStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value for AZURE_OPENAI_TEMPERATURE: %v", err)
	}

	topP, err := strconv.ParseFloat(topPStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value for AZURE_OPENAI_TOP_P: %v", err)
	}

	config := Config{
		APIKey:         apiKey,
		Endpoint:       endpoint,
		DeploymentName: deploymentName,
		APIVersion:     apiVersion,
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		TopP:           topP,
	}

	return NewClient(config), nil
}

// Call makes a call to the LLM API with the given system prompt and user message
func (c *Client) Call(systemPrompt, userMessage string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	return c.ChatCompletion(messages)
}

// ChatCompletion makes a chat completion call to the OpenAI API
func (c *Client) ChatCompletion(messages []Message) (string, error) {
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		c.config.Endpoint, c.config.DeploymentName, c.config.APIVersion)

	// Prepare the request body
	requestBody := ChatRequest{
		Model:            c.config.DeploymentName,
		Messages:         messages,
		MaxTokens:        c.config.MaxTokens,
		Temperature:      c.config.Temperature,
		TopP:             c.config.TopP,
		FrequencyPenalty: c.config.FrequencyPenalty,
		PresencePenalty:  c.config.PresencePenalty,
	}

	reqJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.config.APIKey)

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body with a larger buffer capacity
	// Allow for up to 10MB response
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Unmarshal the response
	var chatResponse ChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Extract the generated content
	if len(chatResponse.Choices) == 0 {
		return "", errors.New("no completions returned from the API")
	}

	// Return the full untruncated message content
	return chatResponse.Choices[0].Message.Content, nil
}
