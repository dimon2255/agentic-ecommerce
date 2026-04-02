package assistant

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/voyage"
)

//go:embed prompts/system.txt
var systemPromptTemplate string

type assistantService struct {
	repo      Repository
	voyage    *voyage.Client
	anthropic *anthropic.Client
}

// NewService creates an assistant service with the given dependencies.
func NewService(repo Repository, voyageClient *voyage.Client, anthropicClient *anthropic.Client) Service {
	return &assistantService{
		repo:      repo,
		voyage:    voyageClient,
		anthropic: anthropicClient,
	}
}

func (s *assistantService) Chat(ctx context.Context, userID string, req ChatRequest) (*ChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Create or reuse session
	sessionID := req.SessionID
	if sessionID == "" {
		title := req.Message
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		session, err := s.repo.CreateSession(ctx, userID, title)
		if err != nil {
			return nil, apperror.NewInternal("failed to create chat session", err)
		}
		sessionID = session.ID
	}

	// Save user message
	_, err := s.repo.SaveMessage(ctx, sessionID, "user", req.Message, nil)
	if err != nil {
		return nil, apperror.NewInternal("failed to save user message", err)
	}

	// Embed user query for semantic search
	embeddings, err := s.voyage.Embed(ctx, []string{req.Message})
	if err != nil {
		return nil, apperror.NewInternal("failed to embed query", err)
	}
	if len(embeddings) == 0 {
		return nil, apperror.NewInternal("no embedding returned", nil)
	}

	// Vector similarity search
	matches, err := s.repo.MatchProducts(ctx, embeddings[0], 0.3, 5)
	if err != nil {
		return nil, apperror.NewInternal("failed to search products", err)
	}

	// Build system prompt with product context
	systemPrompt := buildSystemPrompt(matches)

	// Call Anthropic
	completion, err := s.anthropic.Complete(ctx, anthropic.CompletionRequest{
		System: systemPrompt,
		Messages: []anthropic.Message{
			{Role: "user", Content: req.Message},
		},
		MaxTokens: 1024,
	})
	if err != nil {
		return nil, apperror.NewInternal("failed to generate response", err)
	}

	// Extract product IDs from matches
	productIDs := make([]string, len(matches))
	for i, m := range matches {
		productIDs[i] = m.ProductID
	}

	// Save assistant response
	msg, err := s.repo.SaveMessage(ctx, sessionID, "assistant", completion, productIDs)
	if err != nil {
		return nil, apperror.NewInternal("failed to save assistant response", err)
	}

	return &ChatResponse{
		SessionID: sessionID,
		Message:   *msg,
	}, nil
}

func buildSystemPrompt(matches []ProductMatch) string {
	if len(matches) == 0 {
		return strings.Replace(systemPromptTemplate, "{{PRODUCT_CONTEXT}}", "No products found matching the query.", 1)
	}

	var sb strings.Builder
	for i, m := range matches {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		sb.WriteString(m.Content)
		sb.WriteString(fmt.Sprintf("\n(Relevance: %.0f%%)", m.Similarity*100))
	}

	return strings.Replace(systemPromptTemplate, "{{PRODUCT_CONTEXT}}", sb.String(), 1)
}
