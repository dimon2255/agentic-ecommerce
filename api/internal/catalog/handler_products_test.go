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

func setupTestProductHandler(supabaseHandler http.HandlerFunc) (*ProductHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewProductHandler(client)
	return handler, server
}

func TestListProducts(t *testing.T) {
	handler, server := setupTestProductHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Product{
			{ID: "1", Name: "T-Shirt", Slug: "t-shirt", BasePrice: 29.99, Status: "active"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/products?category_id=cat-1", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []Product
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 product, got %d", len(result))
	}
}

func TestGetProductBySlug(t *testing.T) {
	handler, server := setupTestProductHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Product{ID: "1", Name: "T-Shirt", Slug: "t-shirt", BasePrice: 29.99})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/products/t-shirt", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "t-shirt")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetBySlug(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result Product
	json.NewDecoder(w.Body).Decode(&result)
	if result.Slug != "t-shirt" {
		t.Errorf("expected slug=t-shirt, got %s", result.Slug)
	}
}

func TestCreateProduct(t *testing.T) {
	handler, server := setupTestProductHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]Product{{ID: "1", Name: "T-Shirt", Slug: "t-shirt", BasePrice: 29.99}})
	})
	defer server.Close()

	body := `{"category_id":"cat-1","name":"T-Shirt","slug":"t-shirt","base_price":29.99,"status":"draft"}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
