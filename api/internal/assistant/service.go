package assistant

import "context"

// Service defines business operations for the assistant domain.
type Service interface {
	// Chat processes a user message and returns an AI-generated response.
	Chat(ctx context.Context, userID string, req ChatRequest) (*ChatResponse, error)
}
