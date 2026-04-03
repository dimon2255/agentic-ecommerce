package assistant

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/voyage"
)

//go:embed prompts/system.txt
var systemPromptTemplate string

//go:embed prompts/system_tools.txt
var systemToolsPrompt string

const maxToolIterations = 5
const maxHistoryMessages = 20

type assistantService struct {
	repo         Repository
	voyage       *voyage.Client
	anthropic    *anthropic.Client
	toolExecutor *ToolExecutor
}

// NewService creates an assistant service with the given dependencies.
func NewService(repo Repository, voyageClient *voyage.Client, anthropicClient *anthropic.Client, catalogSvc catalog.Service, cartSvc cart.Service) Service {
	return &assistantService{
		repo:         repo,
		voyage:       voyageClient,
		anthropic:    anthropicClient,
		toolExecutor: NewToolExecutor(catalogSvc, cartSvc),
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
		log.Printf("[assistant] Voyage embed error: %v", err)
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
		log.Printf("[assistant] Anthropic API error: %v", err)
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

func (s *assistantService) ChatWithTools(ctx context.Context, userID string, req ChatRequest) (*ChatResponse, error) {
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

	// Load conversation history and build messages
	messages, err := s.buildMessagesWithHistory(ctx, sessionID, req.Message)
	if err != nil {
		return nil, err
	}

	// Agentic tool loop
	tools := AllTools()
	var cartUpdated bool
	toolsUsed := make(map[string]bool)
	var finalText string

	for i := 0; i < maxToolIterations; i++ {
		resp, err := s.anthropic.CompleteWithTools(ctx, anthropic.ToolCompletionRequest{
			System:    systemToolsPrompt,
			Messages:  messages,
			Tools:     tools,
			MaxTokens: 2048,
		})
		if err != nil {
			log.Printf("[assistant] Anthropic API error (iteration %d): %v", i, err)
			return nil, apperror.NewInternal("failed to generate response", err)
		}

		// If Claude is done talking, extract text and break
		if resp.StopReason == "end_turn" || resp.StopReason == "max_tokens" {
			finalText = resp.TextContent()
			break
		}

		// Tool use — execute each tool call
		if resp.StopReason == "tool_use" {
			// Append assistant message with full content blocks
			messages = append(messages, anthropic.RichMessage{
				Role:    "assistant",
				Content: resp.Content,
			})

			// Execute tools and collect results
			var toolResults []anthropic.ToolResultBlock
			for _, block := range resp.ToolUseBlocks() {
				toolsUsed[block.Name] = true
				result := s.toolExecutor.Execute(ctx, block, userID)
				if result.CartUpdated {
					cartUpdated = true
				}
				toolResults = append(toolResults, anthropic.ToolResultBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result.Content,
					IsError:   result.IsError,
				})
			}

			// Append tool results as user message
			messages = append(messages, anthropic.RichMessage{
				Role:    "user",
				Content: toolResults,
			})
			continue
		}

		// Unknown stop reason — take whatever text we have
		finalText = resp.TextContent()
		break
	}

	if finalText == "" {
		finalText = "I'm sorry, I wasn't able to complete my research. Could you try rephrasing your question?"
	}

	// Save final assistant response
	usedToolNames := make([]string, 0, len(toolsUsed))
	for name := range toolsUsed {
		usedToolNames = append(usedToolNames, name)
	}

	msg, err := s.repo.SaveMessage(ctx, sessionID, "assistant", finalText, nil)
	if err != nil {
		return nil, apperror.NewInternal("failed to save assistant response", err)
	}

	return &ChatResponse{
		SessionID:   sessionID,
		Message:     *msg,
		CartUpdated: cartUpdated,
		ToolsUsed:   usedToolNames,
	}, nil
}

// buildMessagesWithHistory loads conversation history and appends the current message.
func (s *assistantService) buildMessagesWithHistory(ctx context.Context, sessionID, currentMessage string) ([]anthropic.RichMessage, error) {
	history, err := s.repo.GetSessionMessages(ctx, sessionID)
	if err != nil {
		log.Printf("[assistant] Failed to load history: %v", err)
		// Fall back to single message if history load fails
		return []anthropic.RichMessage{
			{Role: "user", Content: currentMessage},
		}, nil
	}

	// Cap history and exclude the message we just saved (it's the last one)
	if len(history) > 0 && history[len(history)-1].Role == "user" && history[len(history)-1].Content == currentMessage {
		history = history[:len(history)-1]
	}

	// Cap at maxHistoryMessages (keeping the most recent)
	if len(history) > maxHistoryMessages {
		history = history[len(history)-maxHistoryMessages:]
	}

	// Convert to RichMessages, ensuring alternation starting with user
	var messages []anthropic.RichMessage
	for _, m := range history {
		// Skip if same role as previous (enforce alternation)
		if len(messages) > 0 && messages[len(messages)-1].Role == m.Role {
			continue
		}
		messages = append(messages, anthropic.RichMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Ensure conversation starts with user role
	if len(messages) > 0 && messages[0].Role != "user" {
		messages = messages[1:]
	}

	// Append current user message
	messages = append(messages, anthropic.RichMessage{
		Role:    "user",
		Content: currentMessage,
	})

	return messages, nil
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
