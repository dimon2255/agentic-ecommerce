package cart

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestCartHandler(supabaseHandler http.HandlerFunc) (*CartHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key", 10*time.Second)
	handler := NewCartHandler(client)
	return handler, server
}

func withUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestGetCart_EmptyForGuest(t *testing.T) {
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {
		// Return empty array — no cart exists
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Cart{})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/cart", nil)
	req.Header.Set("X-Session-ID", "session-abc")
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result CartResponse
	json.NewDecoder(w.Body).Decode(&result)
	if result.ID != "" {
		t.Errorf("expected empty cart ID, got %s", result.ID)
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
}

func TestGetCart_WithItems(t *testing.T) {
	callCount := 0
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			// First call: find cart
			json.NewEncoder(w).Encode([]Cart{
				{ID: "cart-1", SessionID: "session-abc", Status: "active"},
			})
		} else {
			// Second call: cart items with embedded SKU data
			json.NewEncoder(w).Encode([]CartItemWithSKU{
				{
					ID: "item-1", CartID: "cart-1", SKUID: "sku-1",
					Quantity: 2, UnitPrice: 24.99,
					SKU: SKUEmbed{
						SKUCode: "TEE-BLK-M",
						Product: ProductEmbed{Name: "Classic Tee", Slug: "classic-tee", BasePrice: 24.99},
					},
				},
			})
		}
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/cart", nil)
	req.Header.Set("X-Session-ID", "session-abc")
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result CartResponse
	json.NewDecoder(w.Body).Decode(&result)
	if result.ID != "cart-1" {
		t.Errorf("expected cart-1, got %s", result.ID)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].SKU.SKUCode != "TEE-BLK-M" {
		t.Errorf("expected TEE-BLK-M, got %s", result.Items[0].SKU.SKUCode)
	}
}

func TestGetCart_RequiresSessionOrAuth(t *testing.T) {
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	req := httptest.NewRequest("GET", "/cart", nil)
	// No session ID, no auth
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAddItem_CreatesCartAndAddsItem(t *testing.T) {
	callCount := 0
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		switch callCount {
		case 1:
			// Find existing cart — none found
			json.NewEncoder(w).Encode([]Cart{})
		case 2:
			// Create new cart
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]Cart{
				{ID: "cart-new", SessionID: "session-abc", Status: "active"},
			})
		case 3:
			// Look up SKU price
			json.NewEncoder(w).Encode([]SKUForPrice{
				{PriceOverride: nil, Product: ProductEmbed{BasePrice: 24.99}},
			})
		case 4:
			// Check for existing item with same SKU — none
			json.NewEncoder(w).Encode([]CartItem{})
		case 5:
			// Insert cart item
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]CartItem{
				{ID: "item-1", CartID: "cart-new", SKUID: "sku-1", Quantity: 1, UnitPrice: 24.99},
			})
		case 6:
			// Fetch enriched items for response
			json.NewEncoder(w).Encode([]CartItemWithSKU{
				{
					ID: "item-1", CartID: "cart-new", SKUID: "sku-1",
					Quantity: 1, UnitPrice: 24.99,
					SKU: SKUEmbed{
						SKUCode: "TEE-BLK-M",
						Product: ProductEmbed{Name: "Classic Tee", Slug: "classic-tee", BasePrice: 24.99},
					},
				},
			})
		}
	})
	defer server.Close()

	body := `{"sku_id":"sku-1","quantity":1}`
	req := httptest.NewRequest("POST", "/cart/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "session-abc")
	w := httptest.NewRecorder()

	handler.AddItem(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var result CartResponse
	json.NewDecoder(w.Body).Decode(&result)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
}

func TestAddItem_ValidationErrors(t *testing.T) {
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	body := `{"sku_id":"","quantity":0}`
	req := httptest.NewRequest("POST", "/cart/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "session-abc")
	w := httptest.NewRecorder()

	handler.AddItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateItem_ChangesQuantity(t *testing.T) {
	callCount := 0
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		switch callCount {
		case 1:
			// Find cart
			json.NewEncoder(w).Encode([]Cart{
				{ID: "cart-1", SessionID: "session-abc", Status: "active"},
			})
		case 2:
			// Find item in cart
			json.NewEncoder(w).Encode([]CartItem{
				{ID: "item-1", CartID: "cart-1", SKUID: "sku-1", Quantity: 1, UnitPrice: 24.99},
			})
		case 3:
			// Update quantity
			json.NewEncoder(w).Encode([]CartItem{
				{ID: "item-1", CartID: "cart-1", SKUID: "sku-1", Quantity: 3, UnitPrice: 24.99},
			})
		case 4:
			// Fetch enriched items for response
			json.NewEncoder(w).Encode([]CartItemWithSKU{
				{
					ID: "item-1", CartID: "cart-1", SKUID: "sku-1",
					Quantity: 3, UnitPrice: 24.99,
					SKU: SKUEmbed{SKUCode: "TEE-BLK-M", Product: ProductEmbed{Name: "Classic Tee", BasePrice: 24.99}},
				},
			})
		}
	})
	defer server.Close()

	body := `{"quantity":3}`
	req := httptest.NewRequest("PATCH", "/cart/items/item-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "session-abc")
	req = withChiParam(req, "itemId", "item-1")
	w := httptest.NewRecorder()

	handler.UpdateItem(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result CartResponse
	json.NewDecoder(w.Body).Decode(&result)
	if len(result.Items) != 1 || result.Items[0].Quantity != 3 {
		t.Errorf("expected quantity 3, got %+v", result.Items)
	}
}

func TestRemoveItem_DeletesItem(t *testing.T) {
	callCount := 0
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		switch callCount {
		case 1:
			// Find cart
			json.NewEncoder(w).Encode([]Cart{
				{ID: "cart-1", SessionID: "session-abc", Status: "active"},
			})
		case 2:
			// Find item in cart
			json.NewEncoder(w).Encode([]CartItem{
				{ID: "item-1", CartID: "cart-1", SKUID: "sku-1", Quantity: 1, UnitPrice: 24.99},
			})
		case 3:
			// Delete item
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]CartItem{})
		}
	})
	defer server.Close()

	req := httptest.NewRequest("DELETE", "/cart/items/item-1", nil)
	req.Header.Set("X-Session-ID", "session-abc")
	req = withChiParam(req, "itemId", "item-1")
	w := httptest.NewRecorder()

	handler.RemoveItem(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMergeCart_MovesGuestItemsToUserCart(t *testing.T) {
	callCount := 0
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		switch callCount {
		case 1:
			// Find guest cart by session_id
			json.NewEncoder(w).Encode([]Cart{
				{ID: "guest-cart", SessionID: "session-abc", Status: "active"},
			})
		case 2:
			// Find user cart
			json.NewEncoder(w).Encode([]Cart{
				{ID: "user-cart", UserID: strPtr("user-123"), SessionID: "session-old", Status: "active"},
			})
		case 3:
			// Fetch guest cart items
			json.NewEncoder(w).Encode([]CartItem{
				{ID: "guest-item-1", CartID: "guest-cart", SKUID: "sku-1", Quantity: 2, UnitPrice: 24.99},
			})
		case 4:
			// Check for duplicate SKU in user cart — none found
			json.NewEncoder(w).Encode([]CartItem{})
		case 5:
			// Move item: update cart_id
			json.NewEncoder(w).Encode([]CartItem{
				{ID: "guest-item-1", CartID: "user-cart", SKUID: "sku-1", Quantity: 2, UnitPrice: 24.99},
			})
		case 6:
			// Mark guest cart as merged
			json.NewEncoder(w).Encode([]Cart{
				{ID: "guest-cart", SessionID: "session-abc", Status: "merged"},
			})
		case 7:
			// Fetch enriched user cart items for response
			json.NewEncoder(w).Encode([]CartItemWithSKU{
				{
					ID: "guest-item-1", CartID: "user-cart", SKUID: "sku-1",
					Quantity: 2, UnitPrice: 24.99,
					SKU: SKUEmbed{SKUCode: "TEE-BLK-M", Product: ProductEmbed{Name: "Classic Tee", BasePrice: 24.99}},
				},
			})
		}
	})
	defer server.Close()

	body := `{"session_id":"session-abc"}`
	req := httptest.NewRequest("POST", "/cart/merge", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-123")
	w := httptest.NewRecorder()

	handler.MergeCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result CartResponse
	json.NewDecoder(w.Body).Decode(&result)
	if result.ID != "user-cart" {
		t.Errorf("expected user-cart, got %s", result.ID)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
}

func TestMergeCart_RequiresAuth(t *testing.T) {
	handler, server := setupTestCartHandler(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	body := `{"session_id":"session-abc"}`
	req := httptest.NewRequest("POST", "/cart/merge", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No user ID set
	w := httptest.NewRecorder()

	handler.MergeCart(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func strPtr(s string) *string { return &s }
