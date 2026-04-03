package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompleteWithTools_EndTurn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body contains tools
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["tools"]; !ok {
			t.Error("expected tools in request body")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"stop_reason": "end_turn",
			"content": []map[string]any{
				{"type": "text", "text": "Here are some laptops for you."},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "test-model")
	client.SetBaseURL(server.URL)

	resp, err := client.CompleteWithTools(context.Background(), ToolCompletionRequest{
		System: "You are a helpful assistant.",
		Messages: []RichMessage{
			{Role: "user", Content: "show me laptops"},
		},
		Tools: []Tool{
			{Name: "search", Description: "search products", InputSchema: json.RawMessage(`{"type":"object"}`)},
		},
		MaxTokens: 1024,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StopReason != "end_turn" {
		t.Errorf("expected stop_reason=end_turn, got %s", resp.StopReason)
	}
	if text := resp.TextContent(); text != "Here are some laptops for you." {
		t.Errorf("unexpected text: %s", text)
	}
	if len(resp.ToolUseBlocks()) != 0 {
		t.Error("expected no tool_use blocks")
	}
}

func TestCompleteWithTools_ToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"stop_reason": "tool_use",
			"content": []map[string]any{
				{"type": "text", "text": "Let me search for that."},
				{
					"type":  "tool_use",
					"id":    "toolu_01abc",
					"name":  "search_products",
					"input": map[string]any{"query": "laptops"},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "test-model")
	client.SetBaseURL(server.URL)

	resp, err := client.CompleteWithTools(context.Background(), ToolCompletionRequest{
		System: "test",
		Messages: []RichMessage{
			{Role: "user", Content: "find laptops"},
		},
		Tools:     []Tool{{Name: "search_products", Description: "search", InputSchema: json.RawMessage(`{"type":"object"}`)}},
		MaxTokens: 2048,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StopReason != "tool_use" {
		t.Errorf("expected stop_reason=tool_use, got %s", resp.StopReason)
	}

	toolBlocks := resp.ToolUseBlocks()
	if len(toolBlocks) != 1 {
		t.Fatalf("expected 1 tool_use block, got %d", len(toolBlocks))
	}
	if toolBlocks[0].Name != "search_products" {
		t.Errorf("expected tool name=search_products, got %s", toolBlocks[0].Name)
	}
	if toolBlocks[0].ID != "toolu_01abc" {
		t.Errorf("expected tool id=toolu_01abc, got %s", toolBlocks[0].ID)
	}

	text := resp.TextContent()
	if text != "Let me search for that." {
		t.Errorf("unexpected text: %s", text)
	}
}

func TestCompleteWithTools_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"type":    "invalid_request_error",
				"message": "max_tokens must be positive",
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "test-model")
	client.SetBaseURL(server.URL)

	_, err := client.CompleteWithTools(context.Background(), ToolCompletionRequest{
		Messages: []RichMessage{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestCompleteWithTools_DefaultMaxTokens(t *testing.T) {
	var receivedMaxTokens float64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		receivedMaxTokens = body["max_tokens"].(float64)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"stop_reason": "end_turn",
			"content":     []map[string]any{{"type": "text", "text": "ok"}},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "test-model")
	client.SetBaseURL(server.URL)

	client.CompleteWithTools(context.Background(), ToolCompletionRequest{
		Messages: []RichMessage{{Role: "user", Content: "hi"}},
		// MaxTokens not set — should default to 2048
	})

	if receivedMaxTokens != 2048 {
		t.Errorf("expected default max_tokens=2048, got %.0f", receivedMaxTokens)
	}
}

func TestRichMessage_MarshalJSON(t *testing.T) {
	// String content
	msg := RichMessage{Role: "user", Content: "hello"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["role"] != "user" || parsed["content"] != "hello" {
		t.Errorf("unexpected marshal: %s", string(data))
	}

	// Content blocks
	msg2 := RichMessage{Role: "assistant", Content: []ContentBlock{
		{Type: "text", Text: "Let me search."},
		{Type: "tool_use", ID: "t1", Name: "search", Input: json.RawMessage(`{"q":"hi"}`)},
	}}
	data2, err := json.Marshal(msg2)
	if err != nil {
		t.Fatal(err)
	}
	var parsed2 map[string]any
	json.Unmarshal(data2, &parsed2)
	blocks := parsed2["content"].([]any)
	if len(blocks) != 2 {
		t.Errorf("expected 2 content blocks, got %d", len(blocks))
	}

	// Tool result blocks
	msg3 := RichMessage{Role: "user", Content: []ToolResultBlock{
		{Type: "tool_result", ToolUseID: "t1", Content: `{"data":"ok"}`},
	}}
	data3, err := json.Marshal(msg3)
	if err != nil {
		t.Fatal(err)
	}
	var parsed3 map[string]any
	json.Unmarshal(data3, &parsed3)
	results := parsed3["content"].([]any)
	if len(results) != 1 {
		t.Errorf("expected 1 tool result, got %d", len(results))
	}
}
