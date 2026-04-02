package assistant

import (
	"context"
	"fmt"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type supabaseRepository struct {
	db *supabase.Client
}

// NewSupabaseRepository creates an assistant repository backed by Supabase PostgREST.
func NewSupabaseRepository(db *supabase.Client) Repository {
	return &supabaseRepository{db: db}
}

func (r *supabaseRepository) CreateSession(_ context.Context, userID, title string) (*ChatSession, error) {
	var sessions []ChatSession
	err := r.db.From("chat_sessions").Insert(map[string]any{
		"user_id": userID,
		"title":   title,
	}).Execute(&sessions)
	if err != nil {
		return nil, fmt.Errorf("create chat session: %w", err)
	}
	return &sessions[0], nil
}

func (r *supabaseRepository) SaveMessage(_ context.Context, sessionID, role, content string, productIDs []string) (*ChatMessage, error) {
	if productIDs == nil {
		productIDs = []string{}
	}
	var messages []ChatMessage
	err := r.db.From("chat_messages").Insert(map[string]any{
		"session_id":  sessionID,
		"role":        role,
		"content":     content,
		"product_ids": productIDs,
	}).Execute(&messages)
	if err != nil {
		return nil, fmt.Errorf("save chat message: %w", err)
	}
	return &messages[0], nil
}

func (r *supabaseRepository) GetSessionMessages(_ context.Context, sessionID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	err := r.db.From("chat_messages").
		Select("*").
		Eq("session_id", sessionID).
		Order("created_at", "asc").
		Execute(&messages)
	if err != nil {
		return nil, fmt.Errorf("get session messages: %w", err)
	}
	if messages == nil {
		messages = []ChatMessage{}
	}
	return messages, nil
}

func (r *supabaseRepository) MatchProducts(_ context.Context, queryEmbedding []float32, threshold float64, limit int) ([]ProductMatch, error) {
	// Convert float32 slice to string representation for PostgREST
	embeddingStr := Float32SliceToVectorString(queryEmbedding)

	var matches []ProductMatch
	err := r.db.RPC("match_products", map[string]any{
		"query_embedding": embeddingStr,
		"match_threshold": threshold,
		"match_count":     limit,
	}, &matches)
	if err != nil {
		return nil, fmt.Errorf("match products: %w", err)
	}
	if matches == nil {
		matches = []ProductMatch{}
	}
	return matches, nil
}

func (r *supabaseRepository) UpsertEmbedding(_ context.Context, record EmbeddingRecord) error {
	// PostgREST upsert: INSERT with ON CONFLICT via Prefer header
	// Since we have a unique index on product_id, we use the merge approach:
	// Delete existing then insert (PostgREST upsert requires primary key, not unique index)
	err := r.db.From("product_embeddings").
		Eq("product_id", record.ProductID).
		Delete().
		Execute(nil)
	if err != nil {
		// Ignore delete errors (row might not exist)
	}

	return r.db.From("product_embeddings").Insert(map[string]any{
		"product_id": record.ProductID,
		"content":    record.Content,
		"embedding":  record.Embedding,
		"metadata":   record.Metadata,
	}).Execute(nil)
}

// Float32SliceToVectorString converts a float32 slice to pgvector string format: "[0.1,0.2,...]"
func Float32SliceToVectorString(v []float32) string {
	buf := make([]byte, 0, len(v)*10)
	buf = append(buf, '[')
	for i, f := range v {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, fmt.Sprintf("%g", f)...)
	}
	buf = append(buf, ']')
	return string(buf)
}
