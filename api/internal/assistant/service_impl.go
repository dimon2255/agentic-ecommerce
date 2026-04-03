package assistant

import (
	"context"
	_ "embed"
	"encoding/json"
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
	messages, err := s.loadConversationMessages(ctx, sessionID, req.Message)
	if err != nil {
		return nil, err
	}

	// Agentic tool loop
	tools := AllTools()
	var cartUpdated bool
	toolsUsed := make(map[string]bool)
	var finalText string

	for i := range maxToolIterations {
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

// loadConversationMessages loads history from DB and delegates to buildConversationMessages.
func (s *assistantService) loadConversationMessages(ctx context.Context, sessionID, currentMessage string) ([]anthropic.RichMessage, error) {
	history, err := s.repo.GetSessionMessages(ctx, sessionID)
	if err != nil {
		log.Printf("[assistant] Failed to load history: %v", err)
		return []anthropic.RichMessage{
			{Role: "user", Content: currentMessage},
		}, nil
	}
	return buildConversationMessages(history, currentMessage, maxHistoryMessages), nil
}

func (s *assistantService) StreamChat(ctx context.Context, userID string, req ChatRequest, emit func(event, data string)) error {
	if err := req.Validate(); err != nil {
		return err
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
			return apperror.NewInternal("failed to create chat session", err)
		}
		sessionID = session.ID
	}

	// Emit session ID immediately so frontend can track it
	emitJSON(emit, "session", map[string]string{"session_id": sessionID})

	// Save user message
	_, err := s.repo.SaveMessage(ctx, sessionID, "user", req.Message, nil)
	if err != nil {
		return apperror.NewInternal("failed to save user message", err)
	}

	// Load conversation history
	messages, err := s.loadConversationMessages(ctx, sessionID, req.Message)
	if err != nil {
		return err
	}

	// Streaming agentic tool loop
	tools := AllTools()
	var cartUpdated bool
	toolsUsed := make(map[string]bool)
	var finalText strings.Builder

	for i := range maxToolIterations {
		var iterationText strings.Builder

		resp, err := s.anthropic.StreamWithTools(ctx, anthropic.ToolCompletionRequest{
			System:    systemToolsPrompt,
			Messages:  messages,
			Tools:     tools,
			MaxTokens: 2048,
		}, func(event anthropic.StreamEvent) {
			switch event.Type {
			case "content_block_start":
				if event.ContentBlock != nil && event.ContentBlock.Type == "tool_use" {
					emitJSON(emit, "tool_start", map[string]string{"tool": event.ContentBlock.Name})
				}
			case "content_block_delta":
				if event.DeltaType == "text_delta" && event.DeltaText != "" {
					iterationText.WriteString(event.DeltaText)
					emitJSON(emit, "delta", map[string]string{"text": event.DeltaText})
				}
			}
		})
		if err != nil {
			log.Printf("[assistant] Anthropic streaming error (iteration %d): %v", i, err)
			return apperror.NewInternal("failed to generate response", err)
		}

		// If Claude is done — break
		if resp.StopReason == "end_turn" || resp.StopReason == "max_tokens" {
			finalText.WriteString(iterationText.String())
			break
		}

		// Tool use — execute tools and loop
		if resp.StopReason == "tool_use" {
			emitJSON(emit, "status", map[string]string{"status": "thinking"})

			// Append assistant message with full content blocks
			messages = append(messages, anthropic.RichMessage{
				Role:    "assistant",
				Content: resp.Content,
			})

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

			messages = append(messages, anthropic.RichMessage{
				Role:    "user",
				Content: toolResults,
			})
			// Accumulate any text from this iteration
			if iterationText.Len() > 0 {
				finalText.WriteString(iterationText.String())
				// Emit a separator so the next iteration's text doesn't run together
				emitJSON(emit, "delta", map[string]string{"text": "\n\n"})
				finalText.WriteString("\n\n")
			}
			continue
		}

		// Unknown stop reason
		finalText.WriteString(iterationText.String())
		break
	}

	text := finalText.String()
	if text == "" {
		text = "I'm sorry, I wasn't able to complete my research. Could you try rephrasing your question?"
	}

	// Save final assistant response
	_, err = s.repo.SaveMessage(ctx, sessionID, "assistant", text, nil)
	if err != nil {
		log.Printf("[assistant] Failed to save streamed response: %v", err)
	}

	// Emit done event
	usedToolNames := make([]string, 0, len(toolsUsed))
	for name := range toolsUsed {
		usedToolNames = append(usedToolNames, name)
	}
	emitJSON(emit, "done", map[string]any{
		"cart_updated": cartUpdated,
		"tools_used":   usedToolNames,
	})

	return nil
}

// emitJSON marshals data and calls emit with the event type.
func emitJSON(emit func(event, data string), event string, data any) {
	b, _ := json.Marshal(data)
	emit(event, string(b))
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
