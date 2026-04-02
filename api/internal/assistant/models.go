package assistant

import (
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/internal/validate"
)

// ChatRequest is the request body for the chat endpoint.
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

func (r *ChatRequest) Validate() error {
	v := validate.New()
	v.Required("message", r.Message)
	v.MaxLength("message", r.Message, 2000)
	return v.Validate()
}

// ChatResponse is the response body for the chat endpoint.
type ChatResponse struct {
	SessionID string      `json:"session_id"`
	Message   ChatMessage `json:"message"`
}

// ChatSession represents a conversation session.
type ChatSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     *string   `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	ProductIDs []string  `json:"product_ids"`
	CreatedAt  time.Time `json:"created_at"`
}

// ProductMatch represents a product returned from vector similarity search.
type ProductMatch struct {
	ID         string  `json:"id"`
	ProductID  string  `json:"product_id"`
	Content    string  `json:"content"`
	Metadata   any     `json:"metadata"`
	Similarity float64 `json:"similarity"`
}

// EmbeddingRecord is what gets stored in product_embeddings.
type EmbeddingRecord struct {
	ProductID string  `json:"product_id"`
	Content   string  `json:"content"`
	Embedding string  `json:"embedding"`
	Metadata  any     `json:"metadata"`
}
