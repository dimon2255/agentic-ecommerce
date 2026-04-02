package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
func setupTestHandler(
	supabaseHandler http.HandlerFunc,
	voyageHandler http.HandlerFunc,
	anthropicHandler http.HandlerFunc,
) (*Handler, []*httptest.Server) {
	supaServer := httptest.NewServer(supabaseHandler)
	voyageServer := httptest.NewServer(voyageHandler)
	anthropicServer := httptest.NewServer(anthropicHandler)

	db := supa.NewClient(supaServer.URL, "test-key", 10*time.Second)
	repo := NewSupabaseRepository(db)

	vc := voyage.NewClient("test-voyage-key", "voyage-3-large")
	vc.SetBaseURL(voyageServer.URL)

	ac := anthropic.NewClient("test-anthropic-key", "claude-sonnet-4-6-20250514")
	ac.SetBaseURL(anthropicServer.URL)

	svc := NewService(repo, vc, ac)
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
