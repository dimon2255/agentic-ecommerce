package assistant

import "context"

// Service defines business operations for the assistant domain.
type Service interface {
	// Chat processes a user message and returns an AI-generated response (Phase 1 — no tools).
	Chat(ctx context.Context, userID string, req ChatRequest) (*ChatResponse, error)

	// ChatWithTools processes a user message using Claude tool use with an agentic loop.
	ChatWithTools(ctx context.Context, userID string, req ChatRequest) (*ChatResponse, error)

	// StreamChat processes a user message with tool use and streams the response via SSE.
	// The emit callback is called for each SSE event: emit(eventType, jsonData).
	StreamChat(ctx context.Context, userID string, req ChatRequest, emit func(event, data string)) error
}
