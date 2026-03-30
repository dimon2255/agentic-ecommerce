package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestCategoryHandler(supabaseHandler http.HandlerFunc) (*CategoryHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key", 10*time.Second)
	handler := NewCategoryHandler(client)
	return handler, server
}

func TestListCategories(t *testing.T) {
	handler, server := setupTestCategoryHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Category{
			{ID: "1", Name: "Electronics", Slug: "electronics"},
			{ID: "2", Name: "Clothing", Slug: "clothing"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/categories", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []Category
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(result))
	}
}

func TestGetCategoryBySlug(t *testing.T) {
	handler, server := setupTestCategoryHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Category{ID: "1", Name: "Electronics", Slug: "electronics"})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/categories/electronics", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "electronics")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetBySlug(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result Category
	json.NewDecoder(w.Body).Decode(&result)
	if result.Slug != "electronics" {
		t.Errorf("expected slug=electronics, got %s", result.Slug)
	}
}

func TestCreateCategory(t *testing.T) {
	handler, server := setupTestCategoryHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]Category{{ID: "1", Name: "Electronics", Slug: "electronics"}})
	})
	defer server.Close()

	body := `{"name":"Electronics","slug":"electronics"}`
	req := httptest.NewRequest("POST", "/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
