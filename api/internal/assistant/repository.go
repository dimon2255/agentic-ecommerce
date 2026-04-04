package assistant

import "context"

// Repository defines data access operations for the assistant domain.
type Repository interface {
	// CreateSession creates a new chat session for the given user.
	CreateSession(ctx context.Context, userID, title string) (*ChatSession, error)

	// SaveMessage persists a chat message.
	SaveMessage(ctx context.Context, sessionID, role, content string, productIDs []string) (*ChatMessage, error)

	// GetSessionMessages returns messages for a session ordered by creation time.
	GetSessionMessages(ctx context.Context, sessionID string) ([]ChatMessage, error)

	// MatchProducts performs vector similarity search against product embeddings.
	MatchProducts(ctx context.Context, queryEmbedding []float32, threshold float64, limit int) ([]ProductMatch, error)

	// UpsertEmbedding inserts or updates a product embedding.
	UpsertEmbedding(ctx context.Context, record EmbeddingRecord) error

	// SaveTokenUsage persists a token usage record for cost tracking.
	SaveTokenUsage(ctx context.Context, usage TokenUsageRecord) error

	// GetDailyTokenUsage returns aggregated token usage for the current UTC day.
	GetDailyTokenUsage(ctx context.Context) (*DailyTokenUsage, error)
}
