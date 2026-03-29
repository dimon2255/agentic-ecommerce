package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestAttributeHandler(supabaseHandler http.HandlerFunc) (*AttributeHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewAttributeHandler(client)
	return handler, server
}

func TestListAttributes(t *testing.T) {
	handler, server := setupTestAttributeHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]CategoryAttribute{
			{ID: "1", CategoryID: "cat-1", Name: "Size", Type: "enum"},
			{ID: "2", CategoryID: "cat-1", Name: "Color", Type: "enum"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/categories/cat-1/attributes", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("categoryId", "cat-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []CategoryAttribute
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(result))
	}
}

func TestCreateAttribute(t *testing.T) {
	handler, server := setupTestAttributeHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]CategoryAttribute{{ID: "1", CategoryID: "cat-1", Name: "Size", Type: "enum"}})
	})
	defer server.Close()

	body := `{"name":"Size","type":"enum","required":true,"sort_order":0}`
	req := httptest.NewRequest("POST", "/categories/cat-1/attributes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("categoryId", "cat-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
