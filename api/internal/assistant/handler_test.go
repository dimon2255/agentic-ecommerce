package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"
	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/voyage"
)

func withUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

// setupTestHandler creates a handler with mocked Supabase, Voyage, and Anthropic servers.
// catalogSvc and cartSvc can be nil for tests that don't exercise tool use.
func setupTestHandler(
	supabaseHandler http.HandlerFunc,
	voyageHandler http.HandlerFunc,
	anthropicHandler http.HandlerFunc,
) (*Handler, []*httptest.Server) {
	return setupTestHandlerWithServices(supabaseHandler, voyageHandler, anthropicHandler, nil, nil)
}

func setupTestHandlerWithServices(
	supabaseHandler http.HandlerFunc,
	voyageHandler http.HandlerFunc,
	anthropicHandler http.HandlerFunc,
	catalogSvc catalog.Service,
	cartSvc cart.Service,
) (*Handler, []*httptest.Server) {
	supaServer := httptest.NewServer(supabaseHandler)
	voyageServer := httptest.NewServer(voyageHandler)
	anthropicServer := httptest.NewServer(anthropicHandler)

	db := supa.NewClient(supaServer.URL, "test-key", 10*time.Second)
	repo := NewSupabaseRepository(db)

	vc := voyage.NewClient("test-voyage-key", "voyage-3-large")
	vc.SetBaseURL(voyageServer.URL)

	ac := anthropic.NewClient("test-anthropic-key", "claude-sonnet-4-5")
	ac.SetBaseURL(anthropicServer.URL)

	svc := NewService(repo, vc, ac, catalogSvc, cartSvc)
	handler := NewHandler(svc)

	return handler, []*httptest.Server{supaServer, voyageServer, anthropicServer}
}

func closeServers(servers []*httptest.Server) {
	for _, s := range servers {
		s.Close()
	}
}

// dummyEmbedding returns a fake 1024-dim embedding response.
func dummyEmbeddingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		embedding := make([]float32, 1024)
		for i := range embedding {
			embedding[i] = 0.01 * float32(i%100)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": embedding},
			},
		})
	}
}

// dummyAnthropicHandler returns a fixed completion response.
func dummyAnthropicHandler(responseText string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": responseText},
			},
		})
	}
}

func TestChat_Success(t *testing.T) {
	supaCallCount := 0
	supabaseHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		supaCallCount++

		path := r.URL.Path

		// Create session
		if path == "/rest/v1/chat_sessions" && r.Method == "POST" {
			json.NewEncoder(w).Encode([]ChatSession{
				{ID: "session-1", UserID: "user-1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			})
			return
		}

		// Save message (both user and assistant)
		if path == "/rest/v1/chat_messages" && r.Method == "POST" {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			msg := ChatMessage{
				ID:        fmt.Sprintf("msg-%d", supaCallCount),
				SessionID: "session-1",
				Role:      body["role"].(string),
				Content:   body["content"].(string),
				CreatedAt: time.Now(),
			}
			if ids, ok := body["product_ids"].([]any); ok {
				for _, id := range ids {
					msg.ProductIDs = append(msg.ProductIDs, id.(string))
				}
			}
			json.NewEncoder(w).Encode([]ChatMessage{msg})
			return
		}

		// match_products RPC
		if path == "/rest/v1/rpc/match_products" {
			json.NewEncoder(w).Encode([]ProductMatch{
				{
					ID:         "emb-1",
					ProductID:  "prod-1",
					Content:    "Product: ProBook 15\nCategory: Electronics > Laptops\nBase Price: $999.99",
					Similarity: 0.85,
				},
			})
			return
		}

		// Default: empty array
		json.NewEncoder(w).Encode([]any{})
	}

	handler, servers := setupTestHandler(
		supabaseHandler,
		dummyEmbeddingHandler(),
		dummyAnthropicHandler("I'd recommend the ProBook 15 at $999.99."),
	)
	defer closeServers(servers)

	body := `{"message": "I need a laptop for programming"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-1")
	w := httptest.NewRecorder()

	handler.Chat(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ChatResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.SessionID != "session-1" {
		t.Errorf("expected session_id=session-1, got %s", resp.SessionID)
	}
	if resp.Message.Role != "assistant" {
		t.Errorf("expected role=assistant, got %s", resp.Message.Role)
	}
	if !strings.Contains(resp.Message.Content, "ProBook 15") {
		t.Errorf("expected response to mention ProBook 15, got: %s", resp.Message.Content)
	}
}

func TestChat_NoAuth(t *testing.T) {
	handler, servers := setupTestHandler(
		func(w http.ResponseWriter, r *http.Request) {},
		dummyEmbeddingHandler(),
		dummyAnthropicHandler(""),
	)
	defer closeServers(servers)

	body := `{"message": "hello"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No user ID in context
	w := httptest.NewRecorder()

	handler.Chat(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestChat_EmptyMessage(t *testing.T) {
	handler, servers := setupTestHandler(
		func(w http.ResponseWriter, r *http.Request) {},
		dummyEmbeddingHandler(),
		dummyAnthropicHandler(""),
	)
	defer closeServers(servers)

	body := `{"message": ""}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-1")
	w := httptest.NewRecorder()

	handler.Chat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChat_InvalidJSON(t *testing.T) {
	handler, servers := setupTestHandler(
		func(w http.ResponseWriter, r *http.Request) {},
		dummyEmbeddingHandler(),
		dummyAnthropicHandler(""),
	)
	defer closeServers(servers)

	req := httptest.NewRequest("POST", "/", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-1")
	w := httptest.NewRecorder()

	handler.Chat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- ChatWithTools tests ---

// toolUseAnthropicHandler returns tool_use on first call, end_turn text on second.
func toolUseAnthropicHandler(toolName, toolID string, toolInput json.RawMessage, finalText string) http.HandlerFunc {
	var callCount atomic.Int32
	return func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			// First call: respond with tool_use
			json.NewEncoder(w).Encode(map[string]any{
				"stop_reason": "tool_use",
				"content": []map[string]any{
					{"type": "text", "text": "Let me search for that."},
					{"type": "tool_use", "id": toolID, "name": toolName, "input": toolInput},
				},
			})
		} else {
			// Second call: respond with final text
			json.NewEncoder(w).Encode(map[string]any{
				"stop_reason": "end_turn",
				"content": []map[string]any{
					{"type": "text", "text": finalText},
				},
			})
		}
	}
}

// supabaseHandlerForTools handles session + message + history calls (no embedding/match).
func supabaseHandlerForTools() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if path == "/rest/v1/chat_sessions" && r.Method == "POST" {
			json.NewEncoder(w).Encode([]ChatSession{
				{ID: "session-1", UserID: "user-1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			})
			return
		}
		if path == "/rest/v1/chat_messages" && r.Method == "POST" {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			msg := ChatMessage{
				ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
				SessionID: "session-1",
				Role:      body["role"].(string),
				Content:   body["content"].(string),
				CreatedAt: time.Now(),
			}
			json.NewEncoder(w).Encode([]ChatMessage{msg})
			return
		}
		if path == "/rest/v1/chat_messages" && r.Method == "GET" {
			// Return empty history
			json.NewEncoder(w).Encode([]ChatMessage{})
			return
		}
		json.NewEncoder(w).Encode([]any{})
	}
}

// mockCatalogService implements catalog.Service with just the methods used by tools.
type mockCatalogService struct {
	catalog.Service // embed to satisfy interface; unused methods will panic
	products        []catalog.Product
	skus            []catalog.SKU
	categories      []catalog.Category
}

func (m *mockCatalogService) ListProducts(_ context.Context, _ catalog.ProductFilter) ([]catalog.Product, int, error) {
	return m.products, len(m.products), nil
}
func (m *mockCatalogService) GetProductBySlug(_ context.Context, slug string) (*catalog.Product, error) {
	for _, p := range m.products {
		if p.Slug == slug {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockCatalogService) ListSKUsWithAttributes(_ context.Context, _ string) ([]catalog.SKU, error) {
	return m.skus, nil
}
func (m *mockCatalogService) ListCategories(_ context.Context, _ catalog.CategoryFilter) ([]catalog.Category, int, error) {
	return m.categories, len(m.categories), nil
}

// mockCartService implements cart.Service with just the methods used by tools.
type mockCartService struct {
	cart.Service // embed to satisfy interface
	cartResp     *cart.CartResponse
}

func (m *mockCartService) GetCart(_ context.Context, _, _ string) (*cart.CartResponse, error) {
	return m.cartResp, nil
}
func (m *mockCartService) AddItem(_ context.Context, _, _ string, _ cart.AddItemRequest) (*cart.CartResponse, error) {
	return m.cartResp, nil
}

func TestChatWithTools_Success(t *testing.T) {
	catSvc := &mockCatalogService{
		products: []catalog.Product{
			{ID: "p1", Name: "ProBook 15", Slug: "probook-15", BasePrice: 999.99, Status: "active"},
		},
	}
	cartSvc := &mockCartService{
		cartResp: &cart.CartResponse{ID: "cart-1"},
	}

	handler, servers := setupTestHandlerWithServices(
		supabaseHandlerForTools(),
		dummyEmbeddingHandler(), // not used by ChatWithTools, but needed for constructor
		toolUseAnthropicHandler(
			"search_products", "toolu_01", json.RawMessage(`{"query":"laptops"}`),
			"I found the ProBook 15 at $999.99.",
		),
		catSvc,
		cartSvc,
	)
	defer closeServers(servers)

	body := `{"message": "show me laptops"}`
	req := httptest.NewRequest("POST", "/tools", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-1")
	w := httptest.NewRecorder()

	handler.ChatWithTools(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ChatResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.SessionID != "session-1" {
		t.Errorf("expected session-1, got %s", resp.SessionID)
	}
	if !strings.Contains(resp.Message.Content, "ProBook 15") {
		t.Errorf("expected response to contain 'ProBook 15', got: %s", resp.Message.Content)
	}
	if len(resp.ToolsUsed) == 0 {
		t.Error("expected tools_used to be populated")
	}
}

func TestChatWithTools_NoAuth(t *testing.T) {
	handler, servers := setupTestHandlerWithServices(
		supabaseHandlerForTools(),
		dummyEmbeddingHandler(),
		dummyAnthropicHandler(""),
		nil, nil,
	)
	defer closeServers(servers)

	body := `{"message": "hello"}`
	req := httptest.NewRequest("POST", "/tools", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ChatWithTools(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestChatWithTools_MaxIterations(t *testing.T) {
	// Anthropic always returns tool_use — should stop after maxToolIterations
	var callCount atomic.Int32
	alwaysToolUse := func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"stop_reason": "tool_use",
			"content": []map[string]any{
				{"type": "tool_use", "id": "toolu_loop", "name": "get_categories", "input": map[string]any{}},
			},
		})
	}

	catSvc := &mockCatalogService{categories: []catalog.Category{{ID: "c1", Name: "Electronics", Slug: "electronics"}}}
	handler, servers := setupTestHandlerWithServices(
		supabaseHandlerForTools(),
		dummyEmbeddingHandler(),
		alwaysToolUse,
		catSvc,
		&mockCartService{cartResp: &cart.CartResponse{ID: "cart-1"}},
	)
	defer closeServers(servers)

	body := `{"message": "list categories"}`
	req := httptest.NewRequest("POST", "/tools", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-1")
	w := httptest.NewRecorder()

	handler.ChatWithTools(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Should have called Anthropic exactly maxToolIterations (5) times
	if count := callCount.Load(); count != int32(maxToolIterations) {
		t.Errorf("expected %d Anthropic calls, got %d", maxToolIterations, count)
	}
}
