package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ---------------------------------------------------------------------------
// Tool-use types
// ---------------------------------------------------------------------------

// Tool defines a tool that Claude can invoke during conversation.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ContentBlock represents a single block in a message (text or tool_use).
type ContentBlock struct {
	Type  string          `json:"type"`            // "text" or "tool_use"
	Text  string          `json:"text,omitempty"`  // populated when Type == "text"
	ID    string          `json:"id,omitempty"`    // populated when Type == "tool_use"
	Name  string          `json:"name,omitempty"`  // populated when Type == "tool_use"
	Input json.RawMessage `json:"input,omitempty"` // populated when Type == "tool_use"
}

// ToolResultBlock is sent back to Claude with tool execution results.
type ToolResultBlock struct {
	Type      string `json:"type"`                 // always "tool_result"
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// RichMessage supports both simple text content and structured content blocks.
// Content can be: string, []ContentBlock, or []ToolResultBlock.
type RichMessage struct {
	Role    string
	Content any
}

func (m RichMessage) MarshalJSON() ([]byte, error) {
	type wire struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	}
	return json.Marshal(wire{Role: m.Role, Content: m.Content})
}

// ToolCompletionRequest is the request payload for tool-use conversations.
type ToolCompletionRequest struct {
	System    string        `json:"system,omitempty"`
	Messages  []RichMessage `json:"messages"`
	Tools     []Tool        `json:"tools"`
	MaxTokens int           `json:"max_tokens"`
}

// ToolCompletionResponse exposes content blocks and stop_reason.
type ToolCompletionResponse struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
}

// TextContent returns concatenated text from all text blocks.
func (r *ToolCompletionResponse) TextContent() string {
	var text string
	for _, b := range r.Content {
		if b.Type == "text" {
			text += b.Text
		}
	}
	return text
}

// ToolUseBlocks returns only the tool_use content blocks.
func (r *ToolCompletionResponse) ToolUseBlocks() []ContentBlock {
	var blocks []ContentBlock
	for _, b := range r.Content {
		if b.Type == "tool_use" {
			blocks = append(blocks, b)
		}
	}
	return blocks
}

// Client wraps the Anthropic Messages API.
type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates an Anthropic client for chat completions.
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest is the request payload for the Messages API.
type CompletionRequest struct {
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// SetBaseURL overrides the API base URL (used for testing).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

type messagesRequest struct {
	Model     string    `json:"model"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type apiError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends a non-streaming chat completion request and returns the text response.
func (c *Client) Complete(ctx context.Context, req CompletionRequest) (string, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 1024
	}

	body, err := json.Marshal(messagesRequest{
		Model:     c.model,
		System:    req.System,
		Messages:  req.Messages,
		MaxTokens: req.MaxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("anthropic: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("anthropic: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("anthropic: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anthropic: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.Unmarshal(respBody, &apiErr)
		return "", fmt.Errorf("anthropic: API error %d: %s - %s", resp.StatusCode, apiErr.Error.Type, apiErr.Error.Message)
	}

	var result messagesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("anthropic: unmarshal response: %w", err)
	}

	for _, block := range result.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("anthropic: no text content in response")
}

// ---------------------------------------------------------------------------
// Tool-use completion
// ---------------------------------------------------------------------------

type toolMessagesRequest struct {
	Model     string        `json:"model"`
	System    string        `json:"system,omitempty"`
	Messages  []RichMessage `json:"messages"`
	Tools     []Tool        `json:"tools,omitempty"`
	MaxTokens int           `json:"max_tokens"`
}

type toolMessagesResponse struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
}

// CompleteWithTools sends a chat completion that may include tool definitions.
// Unlike Complete, it returns the full response including stop_reason and all
// content block types (text + tool_use).
func (c *Client) CompleteWithTools(ctx context.Context, req ToolCompletionRequest) (*ToolCompletionResponse, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}

	body, err := json.Marshal(toolMessagesRequest{
		Model:     c.model,
		System:    req.System,
		Messages:  req.Messages,
		Tools:     req.Tools,
		MaxTokens: req.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.Unmarshal(respBody, &apiErr)
		return nil, fmt.Errorf("anthropic: API error %d: %s - %s", resp.StatusCode, apiErr.Error.Type, apiErr.Error.Message)
	}

	var result toolMessagesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("anthropic: unmarshal response: %w", err)
	}

	return &ToolCompletionResponse{
		Content:    result.Content,
		StopReason: result.StopReason,
	}, nil
}
