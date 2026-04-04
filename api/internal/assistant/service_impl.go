package assistant

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
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
	repo            Repository
	voyage          *voyage.Client
	anthropic       *anthropic.Client
	toolExecutor    *ToolExecutor
	model           string
	dailyBudgetCents int
}

// ServiceConfig holds optional configuration for the assistant service.
type ServiceConfig struct {
	Model            string
	DailyBudgetCents int
}

// NewService creates an assistant service with the given dependencies.
func NewService(repo Repository, voyageClient *voyage.Client, anthropicClient *anthropic.Client, catalogSvc catalog.Service, cartSvc cart.Service, cfgs ...ServiceConfig) Service {
	svc := &assistantService{
		repo:         repo,
		voyage:       voyageClient,
		anthropic:    anthropicClient,
		toolExecutor: NewToolExecutor(catalogSvc, cartSvc),
	}
	if len(cfgs) > 0 {
		svc.model = cfgs[0].Model
		svc.dailyBudgetCents = cfgs[0].DailyBudgetCents
	}
	return svc
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
		slog.ErrorContext(ctx, "Voyage embed error", "error", err)
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
		slog.ErrorContext(ctx, "Anthropic API error", "error", err)
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

func (s *assistantService) ChatWithTools(ctx context.Context, userID string, isGuest bool, req ChatRequest) (*ChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check daily budget
	if err := s.checkDailyBudget(ctx); err != nil {
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
	if isGuest {
		tools = GuestTools()
	}
	var cartUpdated bool
	toolsUsed := make(map[string]bool)
	var finalText string
	var totalInputTokens, totalOutputTokens int

	for i := range maxToolIterations {
		resp, err := s.anthropic.CompleteWithTools(ctx, anthropic.ToolCompletionRequest{
			System:    systemToolsPrompt,
			Messages:  messages,
			Tools:     tools,
			MaxTokens: 2048,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Anthropic API error", "iteration", i, "error", err)
			return nil, apperror.NewInternal("failed to generate response", err)
		}

		totalInputTokens += resp.Usage.InputTokens
		totalOutputTokens += resp.Usage.OutputTokens

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
				result := s.toolExecutor.Execute(ctx, block, userID, isGuest)
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

	// Persist token usage
	s.saveTokenUsage(ctx, sessionID, totalInputTokens, totalOutputTokens)

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
		slog.ErrorContext(ctx, "failed to load history", "error", err)
		return []anthropic.RichMessage{
			{Role: "user", Content: currentMessage},
		}, nil
	}
	return buildConversationMessages(history, currentMessage, maxHistoryMessages), nil
}

func (s *assistantService) StreamChat(ctx context.Context, userID string, isGuest bool, req ChatRequest, emit func(event, data string)) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// Check daily budget before proceeding
	if err := s.checkDailyBudget(ctx); err != nil {
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
	if isGuest {
		tools = GuestTools()
	}
	var cartUpdated bool
	toolsUsed := make(map[string]bool)
	var finalText strings.Builder
	var totalInputTokens, totalOutputTokens int

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
			slog.ErrorContext(ctx, "Anthropic streaming error", "iteration", i, "error", err)
			return apperror.NewInternal("failed to generate response", err)
		}

		// Accumulate token usage across iterations
		totalInputTokens += resp.Usage.InputTokens
		totalOutputTokens += resp.Usage.OutputTokens

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
				result := s.toolExecutor.Execute(ctx, block, userID, isGuest)
				if result.CartUpdated {
					cartUpdated = true
				}
				// Emit structured tool result to frontend for rich rendering
				if !result.IsError {
					emitJSON(emit, "tool_result", map[string]any{
						"tool":   block.Name,
						"result": json.RawMessage(result.Content),
					})
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
		slog.ErrorContext(ctx, "failed to save streamed response", "error", err)
	}

	// Persist token usage for cost tracking
	s.saveTokenUsage(ctx, sessionID, totalInputTokens, totalOutputTokens)

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

// checkDailyBudget returns an error if the daily token budget has been exceeded.
// Approximate: 1 cent ≈ 1000 tokens (rough average across input/output pricing).
func (s *assistantService) checkDailyBudget(ctx context.Context) error {
	if s.dailyBudgetCents <= 0 {
		return nil // no budget configured
	}
	usage, err := s.repo.GetDailyTokenUsage(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check daily budget", "error", err)
		return nil // fail open — don't block on budget check failure
	}
	totalTokens := usage.TotalInputTokens + usage.TotalOutputTokens
	// Rough approximation: $0.01 per 1000 tokens
	estimatedCents := totalTokens / 1000
	if estimatedCents >= int64(s.dailyBudgetCents) {
		return apperror.NewServiceUnavailable("assistant is temporarily unavailable due to high demand")
	}
	return nil
}

// saveTokenUsage persists accumulated token usage after a conversation turn.
func (s *assistantService) saveTokenUsage(ctx context.Context, sessionID string, inputTokens, outputTokens int) {
	if inputTokens == 0 && outputTokens == 0 {
		return
	}
	model := s.model
	if model == "" {
		model = "unknown"
	}
	if err := s.repo.SaveTokenUsage(ctx, TokenUsageRecord{
		SessionID:    sessionID,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		Model:        model,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to save token usage", "error", err)
	}
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
