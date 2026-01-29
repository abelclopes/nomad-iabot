package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a generic LLM client that supports OpenAI-compatible APIs
// Works with: Ollama, LM Studio, LocalAI, vLLM, text-generation-webui
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
}

// Tool represents a tool/function the LLM can call
type Tool struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a callable function
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Index        int          `json:"index"`
	Message      Message      `json:"message"`
	FinishReason string       `json:"finish_reason"`
	ToolCalls    []ToolCall   `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call from the LLM
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function being called
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a streaming choice
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

// StreamDelta represents the delta content in streaming
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// NewClient creates a new LLM client
func NewClient(baseURL, model string, timeoutSec int) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
	}
}

// Chat sends a chat completion request
func (c *Client) Chat(ctx context.Context, messages []Message, opts ...ChatOption) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
	}

	// Apply options
	for _, opt := range opts {
		opt(&req)
	}

	// Determine endpoint based on provider
	endpoint := c.baseURL + "/v1/chat/completions"
	
	// Ollama uses a different endpoint
	if isOllamaURL(c.baseURL) {
		endpoint = c.baseURL + "/api/chat"
		return c.chatOllama(ctx, messages, opts...)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// chatOllama handles Ollama-specific API format
func (c *Client) chatOllama(ctx context.Context, messages []Message, opts ...ChatOption) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}

	for _, opt := range opts {
		opt(&req)
	}

	// Ollama format
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   false,
		"options": map[string]interface{}{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
		},
	}

	if len(req.Tools) > 0 {
		ollamaReq["tools"] = req.Tools
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.baseURL + "/api/chat"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse Ollama response
	var ollamaResp struct {
		Model     string  `json:"model"`
		Message   Message `json:"message"`
		Done      bool    `json:"done"`
		TotalDuration int64 `json:"total_duration"`
		EvalCount int     `json:"eval_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to standard format
	return &ChatResponse{
		Model: ollamaResp.Model,
		Choices: []Choice{
			{
				Index:        0,
				Message:      ollamaResp.Message,
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			CompletionTokens: ollamaResp.EvalCount,
		},
	}, nil
}

// ChatOption is a function that modifies a ChatRequest
type ChatOption func(*ChatRequest)

// WithMaxTokens sets the max tokens
func WithMaxTokens(n int) ChatOption {
	return func(r *ChatRequest) {
		r.MaxTokens = n
	}
}

// WithTemperature sets the temperature
func WithTemperature(t float64) ChatOption {
	return func(r *ChatRequest) {
		r.Temperature = t
	}
}

// WithStream enables streaming
func WithStream() ChatOption {
	return func(r *ChatRequest) {
		r.Stream = true
	}
}

// WithTools adds tools/functions
func WithTools(tools []Tool) ChatOption {
	return func(r *ChatRequest) {
		r.Tools = tools
	}
}

// ListModels lists available models
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	endpoint := c.baseURL + "/api/tags" // Ollama
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Try OpenAI-compatible endpoint
		endpoint = c.baseURL + "/v1/models"
		httpReq, _ = http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		resp, err = c.httpClient.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to list models: %w", err)
		}
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"models"`
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var models []string
	for _, m := range result.Models {
		models = append(models, m.Name)
	}
	for _, m := range result.Data {
		models = append(models, m.ID)
	}

	return models, nil
}

// Ping checks if the LLM server is reachable
func (c *Client) Ping(ctx context.Context) error {
	endpoint := c.baseURL + "/api/tags" // Ollama
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Try root
		endpoint = c.baseURL
		httpReq, _ = http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		resp, err = c.httpClient.Do(httpReq)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	return nil
}

func isOllamaURL(url string) bool {
	return url == "http://localhost:11434" || 
		   url == "http://127.0.0.1:11434" ||
		   url == "http://host.docker.internal:11434"
}
