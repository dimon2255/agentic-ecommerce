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

func setupTestSKUHandler(supabaseHandler http.HandlerFunc) (*SKUHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewSKUHandler(client)
	return handler, server
}

func TestListSKUs(t *testing.T) {
	callCount := 0
	handler, server := setupTestSKUHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			// First call: list SKUs
			json.NewEncoder(w).Encode([]SKU{
				{ID: "sku-1", ProductID: "prod-1", SKUCode: "TSHIRT-BLU-M", Status: "active"},
			})
		} else {
			// Second call: fetch attribute values
			json.NewEncoder(w).Encode([]SKUAttributeValue{
				{ID: "av-1", SKUID: "sku-1", CategoryAttributeID: "attr-1", Value: "Blue"},
			})
		}
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/products/prod-1/skus", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("productId", "prod-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []SKU
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 SKU, got %d", len(result))
	}
	if result[0].SKUCode != "TSHIRT-BLU-M" {
		t.Errorf("expected SKU code TSHIRT-BLU-M, got %s", result[0].SKUCode)
	}
}

func TestCreateSKU(t *testing.T) {
	callCount := 0
	handler, server := setupTestSKUHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			// Insert SKU
			w.WriteHeader(201)
			json.NewEncoder(w).Encode([]SKU{{ID: "sku-1", ProductID: "prod-1", SKUCode: "TSHIRT-BLU-M"}})
		} else {
			// Insert attribute values
			w.WriteHeader(201)
			json.NewEncoder(w).Encode([]SKUAttributeValue{{ID: "av-1", SKUID: "sku-1"}})
		}
	})
	defer server.Close()

	body := `{"sku_code":"TSHIRT-BLU-M","status":"active","attribute_values":[{"category_attribute_id":"attr-1","value":"Blue"}]}`
	req := httptest.NewRequest("POST", "/products/prod-1/skus", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("productId", "prod-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
