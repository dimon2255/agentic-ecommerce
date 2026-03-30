# Checkout & Stripe Payments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Stripe-powered checkout with 3D Secure support, creating orders from cart items with price re-validation and webhook-driven payment confirmation.

**Architecture:** Frontend collects shipping info and calls Go API to create a draft order + Stripe PaymentIntent. Stripe Payment Element handles card input and 3DS. Stripe webhooks update order status to "paid" and expire the cart. Frontend polls for confirmation.

**Tech Stack:** Nuxt 3, Vue 3, Tailwind CSS, @stripe/stripe-js, Go 1.22+, chi router, stripe-go/v82, Supabase (PostgREST, Auth), PostgreSQL

---

## File Map

### Go API (`api/`)

| File | Responsibility |
|------|---------------|
| `api/pkg/stripe/stripe.go` | Stripe client wrapper (PaymentIntent creation, webhook verification) |
| `api/internal/checkout/payment.go` | PaymentService interface for testability |
| `api/internal/checkout/models.go` | Order, OrderItem, request/response types, internal cart types |
| `api/internal/checkout/handler.go` | StartCheckout, HandleWebhook, GetOrder handlers |
| `api/internal/checkout/handler_test.go` | Handler tests with mock Supabase + mock PaymentService |
| `api/cmd/server/main.go` | Updated: Stripe env vars, checkout handler, routes |

### Nuxt Frontend (`frontend/`)

| File | Responsibility |
|------|---------------|
| `frontend/composables/useCheckout.ts` | Checkout API calls + Stripe Elements lifecycle |
| `frontend/pages/checkout.vue` | Two-step checkout: shipping form → Stripe Payment Element (SPA) |
| `frontend/pages/order/[id].vue` | Order confirmation with status polling (SPA) |
| `frontend/pages/cart.vue` | Updated: working "Proceed to Checkout" button |
| `frontend/nuxt.config.ts` | Updated: order route rule, Stripe publishable key |

### Supabase (`supabase/`)

| File | Responsibility |
|------|---------------|
| `supabase/migrations/00004_orders.sql` | Orders and order_items tables with RLS policies |

---

## Task 1: Database Migration — Orders Schema

**Files:**
- Create: `supabase/migrations/00004_orders.sql`

- [ ] **Step 1: Write the migration file**

```sql
-- Create order status enum
CREATE TYPE order_status AS ENUM ('draft', 'pending', 'paid', 'shipped', 'completed', 'cancelled');

-- Orders table
CREATE TABLE orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES auth.users(id) ON DELETE SET NULL,
    status order_status NOT NULL DEFAULT 'draft',
    email text NOT NULL,
    shipping_address jsonb NOT NULL,
    subtotal numeric(10,2) NOT NULL CHECK (subtotal >= 0),
    total numeric(10,2) NOT NULL CHECK (total >= 0),
    stripe_payment_intent_id text UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Order items table (all fields snapshotted at checkout time)
CREATE TABLE order_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sku_id uuid NOT NULL REFERENCES skus(id) ON DELETE RESTRICT,
    product_name text NOT NULL,
    sku_code text NOT NULL,
    quantity integer NOT NULL CHECK (quantity > 0),
    unit_price numeric(10,2) NOT NULL CHECK (unit_price >= 0)
);

-- Indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_stripe_pi ON orders(stripe_payment_intent_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- RLS
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own orders" ON orders
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Service role full access on orders" ON orders
    FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "Users can view own order items" ON order_items
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM orders WHERE orders.id = order_items.order_id AND orders.user_id = auth.uid())
    );

CREATE POLICY "Service role full access on order_items" ON order_items
    FOR ALL USING (auth.role() = 'service_role');

-- Trigger for updated_at
CREATE TRIGGER orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

- [ ] **Step 2: Apply and verify migration**

Run: `cd supabase && supabase db reset`
Expected: Migration applies without errors. Tables `orders` and `order_items` exist.

- [ ] **Step 3: Commit**

```bash
git add supabase/migrations/00004_orders.sql
git commit -m "feat: add orders and order_items migration"
```

---

## Task 2: Go — Stripe Package + Payment Interface

**Files:**
- Create: `api/internal/checkout/payment.go`
- Create: `api/pkg/stripe/stripe.go`

- [ ] **Step 1: Add stripe-go dependency**

Run: `cd api && go get github.com/stripe/stripe-go/v82`
Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 2: Create PaymentService interface**

This interface lives in the checkout package so the handler depends on an abstraction, not the concrete Stripe client. Tests mock this interface.

```go
// api/internal/checkout/payment.go
package checkout

// PaymentService abstracts payment provider operations for testability.
type PaymentService interface {
	// CreatePaymentIntent creates a payment intent and returns the client secret and payment intent ID.
	CreatePaymentIntent(amountCents int64, currency, orderID string) (clientSecret, paymentIntentID string, err error)
	// VerifyWebhook verifies a webhook signature and returns the event type and payment intent ID.
	VerifyWebhook(payload []byte, sigHeader string) (eventType, paymentIntentID string, err error)
}
```

- [ ] **Step 3: Create Stripe client wrapper implementing PaymentService**

```go
// api/pkg/stripe/stripe.go
package stripe

import (
	"encoding/json"
	"fmt"

	gostripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
)

// Client wraps the Stripe API. Implements checkout.PaymentService.
type Client struct {
	webhookSecret string
}

// NewClient sets the Stripe API key globally and returns a client.
func NewClient(secretKey, webhookSecret string) *Client {
	gostripe.Key = secretKey
	return &Client{webhookSecret: webhookSecret}
}

func (c *Client) CreatePaymentIntent(amountCents int64, currency, orderID string) (string, string, error) {
	params := &gostripe.PaymentIntentParams{
		Amount:             gostripe.Int64(amountCents),
		Currency:           gostripe.String(currency),
		PaymentMethodTypes: gostripe.StringSlice([]string{"card"}),
	}
	params.AddMetadata("order_id", orderID)

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", "", fmt.Errorf("create payment intent: %w", err)
	}
	return pi.ClientSecret, pi.ID, nil
}

func (c *Client) VerifyWebhook(payload []byte, sigHeader string) (string, string, error) {
	event, err := webhook.ConstructEvent(payload, sigHeader, c.webhookSecret)
	if err != nil {
		return "", "", fmt.Errorf("verify webhook signature: %w", err)
	}

	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(event.Data.Raw, &data); err != nil {
		return string(event.Type), "", nil
	}
	return string(event.Type), data.ID, nil
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd api && go build ./...`
Expected: No errors.

- [ ] **Step 5: Commit**

```bash
git add api/internal/checkout/payment.go api/pkg/stripe/stripe.go api/go.mod api/go.sum
git commit -m "feat: add Stripe client wrapper and PaymentService interface"
```

---

## Task 3: Go — Checkout Models

**Files:**
- Create: `api/internal/checkout/models.go`

- [ ] **Step 1: Write model types**

```go
// api/internal/checkout/models.go
package checkout

import "time"

// Database models

type Order struct {
	ID                    string    `json:"id"`
	UserID                *string   `json:"user_id"`
	Status                string    `json:"status"`
	Email                 string    `json:"email"`
	ShippingAddress       any       `json:"shipping_address"`
	Subtotal              float64   `json:"subtotal"`
	Total                 float64   `json:"total"`
	StripePaymentIntentID *string   `json:"stripe_payment_intent_id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	SKUID       string  `json:"sku_id"`
	ProductName string  `json:"product_name"`
	SKUCode     string  `json:"sku_code"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// Request/Response types

type ShippingAddress struct {
	Name    string `json:"name"`
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state,omitempty"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

type StartCheckoutRequest struct {
	Email           string          `json:"email"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
}

type StartCheckoutResponse struct {
	OrderID      string `json:"order_id"`
	ClientSecret string `json:"client_secret"`
}

type PriceChange struct {
	SKUID    string  `json:"sku_id"`
	SKUCode  string  `json:"sku_code"`
	OldPrice float64 `json:"old_price"`
	NewPrice float64 `json:"new_price"`
}

type OrderResponse struct {
	ID              string              `json:"id"`
	Status          string              `json:"status"`
	Email           string              `json:"email"`
	ShippingAddress any                 `json:"shipping_address"`
	Subtotal        float64             `json:"subtotal"`
	Total           float64             `json:"total"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
}

type OrderItemResponse struct {
	ProductName string  `json:"product_name"`
	SKUCode     string  `json:"sku_code"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// Internal types for cart data deserialization (PostgREST embedded select)

type cartItem struct {
	ID        string   `json:"id"`
	SKUID     string   `json:"sku_id"`
	Quantity  int      `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	SKU       skuEmbed `json:"skus"`
}

type skuEmbed struct {
	SKUCode       string       `json:"sku_code"`
	PriceOverride *float64     `json:"price_override"`
	Product       productEmbed `json:"products"`
}

type productEmbed struct {
	Name      string  `json:"name"`
	BasePrice float64 `json:"base_price"`
}

type cart struct {
	ID     string  `json:"id"`
	UserID *string `json:"user_id"`
	Status string  `json:"status"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd api && go build ./...`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add api/internal/checkout/models.go
git commit -m "feat: add checkout domain models and request/response types"
```

---

## Task 4: Go — StartCheckout Handler (TDD)

**Files:**
- Create: `api/internal/checkout/handler.go`
- Create: `api/internal/checkout/handler_test.go`

- [ ] **Step 1: Write test file with mock and happy-path test**

```go
// api/internal/checkout/handler_test.go
package checkout

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

// mockPayments implements PaymentService for tests.
type mockPayments struct {
	clientSecret    string
	paymentIntentID string
	createErr       error
	eventType       string
	webhookPIID     string
	webhookErr      error
}

func (m *mockPayments) CreatePaymentIntent(amountCents int64, currency, orderID string) (string, string, error) {
	return m.clientSecret, m.paymentIntentID, m.createErr
}

func (m *mockPayments) VerifyWebhook(payload []byte, sigHeader string) (string, string, error) {
	return m.eventType, m.webhookPIID, m.webhookErr
}

// newTestHandler creates a CheckoutHandler backed by a mock Supabase server.
func newTestHandler(t *testing.T, mux *http.ServeMux) (*CheckoutHandler, *httptest.Server) {
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	db := supabase.NewClient(server.URL, "test-key")
	payments := &mockPayments{
		clientSecret:    "pi_test_secret_123",
		paymentIntentID: "pi_test_id_123",
	}
	return NewCheckoutHandler(db, payments), server
}

func TestStartCheckout_Success(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/v1/carts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "cart-1", "user_id": "user-1", "status": "active"},
		})
	})

	mux.HandleFunc("/rest/v1/cart_items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": "item-1", "sku_id": "sku-1", "quantity": 2, "unit_price": 29.99,
				"skus": map[string]any{
					"sku_code": "SHIRT-BLU-M", "price_override": nil,
					"products": map[string]any{"name": "Blue Shirt", "base_price": 29.99},
				},
			},
		})
	})

	mux.HandleFunc("/rest/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": "order-1", "status": "draft"},
			})
		case "PATCH":
			w.WriteHeader(http.StatusOK)
		}
	})

	mux.HandleFunc("/rest/v1/order_items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler, _ := newTestHandler(t, mux)

	body := `{"email":"test@example.com","shipping_address":{"name":"John Doe","line1":"123 Main St","city":"Springfield","state":"IL","zip":"62701","country":"US"}}`
	req := httptest.NewRequest("POST", "/checkout/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.StartCheckout(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp StartCheckoutResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.OrderID != "order-1" {
		t.Errorf("expected order_id 'order-1', got %q", resp.OrderID)
	}
	if resp.ClientSecret != "pi_test_secret_123" {
		t.Errorf("expected client_secret 'pi_test_secret_123', got %q", resp.ClientSecret)
	}
}

func TestStartCheckout_EmptyCart(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/v1/carts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "cart-1", "user_id": "user-1", "status": "active"},
		})
	})

	mux.HandleFunc("/rest/v1/cart_items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	})

	handler, _ := newTestHandler(t, mux)

	body := `{"email":"test@example.com","shipping_address":{"name":"John Doe","line1":"123 Main St","city":"Springfield","zip":"62701","country":"US"}}`
	req := httptest.NewRequest("POST", "/checkout/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.StartCheckout(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestStartCheckout_PriceChanged(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/v1/carts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "cart-1", "user_id": "user-1", "status": "active"},
		})
	})

	mux.HandleFunc("/rest/v1/cart_items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Cart snapshot price is 29.99, but current SKU base_price is 34.99
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": "item-1", "sku_id": "sku-1", "quantity": 1, "unit_price": 29.99,
				"skus": map[string]any{
					"sku_code": "SHIRT-BLU-M", "price_override": nil,
					"products": map[string]any{"name": "Blue Shirt", "base_price": 34.99},
				},
			},
		})
	})

	handler, _ := newTestHandler(t, mux)

	body := `{"email":"test@example.com","shipping_address":{"name":"John Doe","line1":"123 Main St","city":"Springfield","zip":"62701","country":"US"}}`
	req := httptest.NewRequest("POST", "/checkout/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.StartCheckout(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	changes, ok := result["price_changes"].([]any)
	if !ok || len(changes) == 0 {
		t.Fatal("expected price_changes in response")
	}
}

func TestStartCheckout_MissingEmail(t *testing.T) {
	handler := &CheckoutHandler{}
	body := `{"email":"","shipping_address":{"name":"John","line1":"123 Main","city":"X","zip":"12345","country":"US"}}`
	req := httptest.NewRequest("POST", "/checkout/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.StartCheckout(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd api && go test ./internal/checkout/ -v -run TestStartCheckout`
Expected: FAIL — `NewCheckoutHandler` and `StartCheckout` not defined.

- [ ] **Step 3: Implement the handler**

```go
// api/internal/checkout/handler.go
package checkout

import (
	"encoding/json"
	"io"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CheckoutHandler struct {
	db       *supabase.Client
	payments PaymentService
}

func NewCheckoutHandler(db *supabase.Client, payments PaymentService) *CheckoutHandler {
	return &CheckoutHandler{db: db, payments: payments}
}

// Routes returns a chi.Router with checkout and order routes.
func (h *CheckoutHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/start", h.StartCheckout)
	return r
}

// OrderRoutes returns a chi.Router for order read endpoints.
func (h *CheckoutHandler) OrderRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", h.GetOrder)
	return r
}

// WebhookRoutes returns a chi.Router for Stripe webhook (no auth middleware).
func (h *CheckoutHandler) WebhookRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.HandleWebhook)
	return r
}

func (h *CheckoutHandler) StartCheckout(w http.ResponseWriter, r *http.Request) {
	var req StartCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		response.Error(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.ShippingAddress.Name == "" || req.ShippingAddress.Line1 == "" ||
		req.ShippingAddress.City == "" || req.ShippingAddress.Zip == "" ||
		req.ShippingAddress.Country == "" {
		response.Error(w, http.StatusBadRequest, "shipping address fields required: name, line1, city, zip, country")
		return
	}

	// Find active cart
	activeCart := h.findActiveCart(r)
	if activeCart == nil {
		response.Error(w, http.StatusBadRequest, "no active cart found")
		return
	}

	// Fetch cart items with embedded SKU + product data
	var items []cartItem
	err := h.db.From("cart_items").
		Select("*,skus(sku_code,price_override,products(name,base_price))").
		Eq("cart_id", activeCart.ID).
		Execute(&items)
	if err != nil || len(items) == 0 {
		response.Error(w, http.StatusBadRequest, "cart is empty")
		return
	}

	// Re-validate prices against current SKU prices
	var priceChanges []PriceChange
	for _, item := range items {
		currentPrice := item.SKU.Product.BasePrice
		if item.SKU.PriceOverride != nil {
			currentPrice = *item.SKU.PriceOverride
		}
		if currentPrice != item.UnitPrice {
			priceChanges = append(priceChanges, PriceChange{
				SKUID:    item.SKUID,
				SKUCode:  item.SKU.SKUCode,
				OldPrice: item.UnitPrice,
				NewPrice: currentPrice,
			})
			// Update cart item to current price
			h.db.From("cart_items").
				Update(map[string]any{"unit_price": currentPrice}).
				Eq("id", item.ID).
				Execute(nil)
		}
	}
	if len(priceChanges) > 0 {
		response.JSON(w, http.StatusConflict, map[string]any{
			"error":         "prices have changed",
			"price_changes": priceChanges,
		})
		return
	}

	// Calculate totals
	var subtotal float64
	for _, item := range items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}
	total := subtotal // No tax/shipping in v1

	// Create draft order
	userID, _ := middleware.GetUserID(r.Context())
	orderData := map[string]any{
		"email":            req.Email,
		"shipping_address": req.ShippingAddress,
		"subtotal":         subtotal,
		"total":            total,
		"status":           "draft",
	}
	if userID != "" {
		orderData["user_id"] = userID
	}

	var orders []Order
	if err := h.db.From("orders").Insert(orderData).Execute(&orders); err != nil || len(orders) == 0 {
		response.Error(w, http.StatusInternalServerError, "failed to create order")
		return
	}
	order := orders[0]

	// Create order items (snapshot product_name and sku_code)
	for _, item := range items {
		orderItemData := map[string]any{
			"order_id":     order.ID,
			"sku_id":       item.SKUID,
			"product_name": item.SKU.Product.Name,
			"sku_code":     item.SKU.SKUCode,
			"quantity":     item.Quantity,
			"unit_price":   item.UnitPrice,
		}
		if err := h.db.From("order_items").Insert(orderItemData).Execute(nil); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to create order items")
			return
		}
	}

	// Create Stripe PaymentIntent
	amountCents := int64(math.Round(total * 100))
	clientSecret, piID, err := h.payments.CreatePaymentIntent(amountCents, "usd", order.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create payment intent")
		return
	}

	// Store payment intent ID on order
	h.db.From("orders").
		Update(map[string]any{"stripe_payment_intent_id": piID}).
		Eq("id", order.ID).
		Execute(nil)

	response.JSON(w, http.StatusOK, StartCheckoutResponse{
		OrderID:      order.ID,
		ClientSecret: clientSecret,
	})
}

// GetOrder and HandleWebhook are implemented in subsequent tasks (stubs for compilation).
func (h *CheckoutHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not implemented")
}

func (h *CheckoutHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	_ = io.ReadAll // prevent unused import
	response.Error(w, http.StatusNotImplemented, "not implemented")
}

// findActiveCart looks up the user's or guest's active cart.
func (h *CheckoutHandler) findActiveCart(r *http.Request) *cart {
	userID, ok := middleware.GetUserID(r.Context())
	if ok {
		var carts []cart
		h.db.From("carts").Select("*").Eq("user_id", userID).Eq("status", "active").Limit(1).Execute(&carts)
		if len(carts) > 0 {
			return &carts[0]
		}
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID != "" {
		var carts []cart
		h.db.From("carts").Select("*").Eq("session_id", sessionID).Eq("status", "active").Is("user_id", "null").Limit(1).Execute(&carts)
		if len(carts) > 0 {
			return &carts[0]
		}
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd api && go test ./internal/checkout/ -v -run TestStartCheckout`
Expected: All 4 `TestStartCheckout_*` tests PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/checkout/handler.go api/internal/checkout/handler_test.go
git commit -m "feat: add StartCheckout handler with price re-validation (TDD)"
```

---

## Task 5: Go — Webhook Handler (TDD)

**Files:**
- Modify: `api/internal/checkout/handler_test.go`
- Modify: `api/internal/checkout/handler.go`

- [ ] **Step 1: Write webhook tests**

Append to `api/internal/checkout/handler_test.go`:

```go
func TestHandleWebhook_PaymentSucceeded(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PATCH":
			w.WriteHeader(http.StatusOK)
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": "order-1", "user_id": "user-1", "stripe_payment_intent_id": "pi_123"},
			})
		}
	})

	mux.HandleFunc("/rest/v1/carts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	db := supabase.NewClient(server.URL, "test-key")
	payments := &mockPayments{
		eventType:   "payment_intent.succeeded",
		webhookPIID: "pi_123",
	}
	handler := NewCheckoutHandler(db, payments)

	req := httptest.NewRequest("POST", "/stripe/webhook", strings.NewReader(`{}`))
	req.Header.Set("Stripe-Signature", "valid-sig")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleWebhook_PaymentFailed(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	db := supabase.NewClient(server.URL, "test-key")
	payments := &mockPayments{
		eventType:   "payment_intent.payment_failed",
		webhookPIID: "pi_456",
	}
	handler := NewCheckoutHandler(db, payments)

	req := httptest.NewRequest("POST", "/stripe/webhook", strings.NewReader(`{}`))
	req.Header.Set("Stripe-Signature", "valid-sig")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	handler := &CheckoutHandler{
		payments: &mockPayments{webhookErr: fmt.Errorf("invalid signature")},
	}

	req := httptest.NewRequest("POST", "/stripe/webhook", strings.NewReader(`{}`))
	req.Header.Set("Stripe-Signature", "bad-sig")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
```

Add `"fmt"` to the import block in the test file.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd api && go test ./internal/checkout/ -v -run TestHandleWebhook`
Expected: FAIL — `HandleWebhook` returns 501 (stub).

- [ ] **Step 3: Replace HandleWebhook stub with real implementation**

Replace the `HandleWebhook` method in `api/internal/checkout/handler.go`:

```go
func (h *CheckoutHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	eventType, piID, err := h.payments.VerifyWebhook(payload, sigHeader)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid webhook signature")
		return
	}

	switch eventType {
	case "payment_intent.succeeded":
		// Update order to paid
		h.db.From("orders").
			Update(map[string]any{"status": "paid"}).
			Eq("stripe_payment_intent_id", piID).
			Execute(nil)

		// Expire the user's active cart (authenticated orders only)
		var orders []Order
		h.db.From("orders").Select("user_id").Eq("stripe_payment_intent_id", piID).Execute(&orders)
		if len(orders) > 0 && orders[0].UserID != nil {
			h.db.From("carts").
				Update(map[string]any{"status": "expired"}).
				Eq("user_id", *orders[0].UserID).
				Eq("status", "active").
				Execute(nil)
		}

	case "payment_intent.payment_failed":
		h.db.From("orders").
			Update(map[string]any{"status": "cancelled"}).
			Eq("stripe_payment_intent_id", piID).
			Execute(nil)
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

Remove the `_ = io.ReadAll` line from the old stub.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd api && go test ./internal/checkout/ -v -run TestHandleWebhook`
Expected: All 3 `TestHandleWebhook_*` tests PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/checkout/handler.go api/internal/checkout/handler_test.go
git commit -m "feat: add Stripe webhook handler for payment confirmation (TDD)"
```

---

## Task 6: Go — GetOrder Handler (TDD)

**Files:**
- Modify: `api/internal/checkout/handler_test.go`
- Modify: `api/internal/checkout/handler.go`

- [ ] **Step 1: Write GetOrder tests**

Append to `api/internal/checkout/handler_test.go`:

```go
func TestGetOrder_Success(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": "order-1", "status": "paid", "email": "test@example.com",
				"shipping_address": map[string]any{"name": "John", "line1": "123 Main", "city": "X", "zip": "12345", "country": "US"},
				"subtotal": 59.98, "total": 59.98,
				"created_at": "2026-03-29T12:00:00Z", "updated_at": "2026-03-29T12:00:00Z",
			},
		})
	})

	mux.HandleFunc("/rest/v1/order_items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "oi-1", "order_id": "order-1", "sku_id": "sku-1", "product_name": "Blue Shirt", "sku_code": "SHIRT-BLU-M", "quantity": 2, "unit_price": 29.99},
		})
	})

	handler, _ := newTestHandler(t, mux)

	// Use chi context to inject URL param
	req := httptest.NewRequest("GET", "/orders/order-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "order-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp OrderResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ID != "order-1" {
		t.Errorf("expected order id 'order-1', got %q", resp.ID)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 order item, got %d", len(resp.Items))
	}
	if resp.Items[0].ProductName != "Blue Shirt" {
		t.Errorf("expected product name 'Blue Shirt', got %q", resp.Items[0].ProductName)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	})

	handler, _ := newTestHandler(t, mux)

	req := httptest.NewRequest("GET", "/orders/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetOrder(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd api && go test ./internal/checkout/ -v -run TestGetOrder`
Expected: FAIL — `GetOrder` returns 501 (stub).

- [ ] **Step 3: Replace GetOrder stub with real implementation**

Replace the `GetOrder` method in `api/internal/checkout/handler.go`:

```go
func (h *CheckoutHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	var orders []Order
	err := h.db.From("orders").Select("*").Eq("id", orderID).Limit(1).Execute(&orders)
	if err != nil || len(orders) == 0 {
		response.Error(w, http.StatusNotFound, "order not found")
		return
	}
	order := orders[0]

	var items []OrderItem
	h.db.From("order_items").Select("*").Eq("order_id", orderID).Execute(&items)
	if items == nil {
		items = []OrderItem{}
	}

	itemResponses := make([]OrderItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = OrderItemResponse{
			ProductName: item.ProductName,
			SKUCode:     item.SKUCode,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	response.JSON(w, http.StatusOK, OrderResponse{
		ID:              order.ID,
		Status:          order.Status,
		Email:           order.Email,
		ShippingAddress: order.ShippingAddress,
		Subtotal:        order.Subtotal,
		Total:           order.Total,
		Items:           itemResponses,
		CreatedAt:       order.CreatedAt,
	})
}
```

Also remove the `_ = io.ReadAll` line if it was leftover from the stub, and clean up the unused `io` import if `HandleWebhook` now uses it properly.

- [ ] **Step 4: Run all checkout tests**

Run: `cd api && go test ./internal/checkout/ -v`
Expected: All tests PASS (StartCheckout, Webhook, GetOrder).

- [ ] **Step 5: Commit**

```bash
git add api/internal/checkout/handler.go api/internal/checkout/handler_test.go
git commit -m "feat: add GetOrder handler for order confirmation (TDD)"
```

---

## Task 7: Go — Wire Routes in main.go

**Files:**
- Modify: `api/cmd/server/main.go`

- [ ] **Step 1: Add Stripe env vars and create checkout handler**

Add these environment variable reads near the existing env var block:

```go
stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
```

Add the import for the stripe package:

```go
stripeClient "github.com/dimon2255/agentic-ecommerce/api/pkg/stripe"
"github.com/dimon2255/agentic-ecommerce/api/internal/checkout"
```

After the existing handler initializations (cartHandler, etc.), add:

```go
stripePayments := stripeClient.NewClient(stripeSecretKey, stripeWebhookSecret)
checkoutHandler := checkout.NewCheckoutHandler(db, stripePayments)
```

- [ ] **Step 2: Register checkout routes**

Inside the `r.Route("/api/v1", ...)` block, after the existing cart routes, add:

```go
r.Route("/checkout", func(r chi.Router) {
	r.Use(auth.OptionalAuth)
	r.Mount("/", checkoutHandler.Routes())
})
r.Route("/orders", func(r chi.Router) {
	r.Use(auth.OptionalAuth)
	r.Mount("/", checkoutHandler.OrderRoutes())
})
```

Outside the `/api/v1` route group (webhooks don't need the API prefix or auth), add:

```go
r.Mount("/stripe/webhook", checkoutHandler.WebhookRoutes())
```

- [ ] **Step 3: Verify compilation**

Run: `cd api && go build ./cmd/server/`
Expected: Compiles with no errors.

- [ ] **Step 4: Commit**

```bash
git add api/cmd/server/main.go
git commit -m "feat: wire checkout, order, and webhook routes in main.go"
```

---

## Task 8: Frontend — Stripe Setup + useCheckout Composable

**Files:**
- Modify: `frontend/package.json` (via npm install)
- Modify: `frontend/nuxt.config.ts`
- Create: `frontend/composables/useCheckout.ts`

- [ ] **Step 1: Install @stripe/stripe-js**

Run: `cd frontend && npm install @stripe/stripe-js`
Expected: Package added to `dependencies` in `package.json`.

- [ ] **Step 2: Update nuxt.config.ts**

Add `stripeKey` to the `runtimeConfig.public` block:

```typescript
runtimeConfig: {
  public: {
    apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9090',
    stripeKey: process.env.NUXT_PUBLIC_STRIPE_KEY || '',
  },
},
```

Add order route rule to the `routeRules` block:

```typescript
'/order/**': { ssr: false },
```

- [ ] **Step 3: Create useCheckout composable**

```typescript
// frontend/composables/useCheckout.ts
import { loadStripe, type Stripe, type StripeElements } from '@stripe/stripe-js'

interface PriceChange {
  sku_id: string
  sku_code: string
  old_price: number
  new_price: number
}

interface StartCheckoutResponse {
  order_id: string
  client_secret: string
}

export interface OrderResponse {
  id: string
  status: string
  email: string
  shipping_address: any
  subtotal: number
  total: number
  items: Array<{
    product_name: string
    sku_code: string
    quantity: number
    unit_price: number
  }>
  created_at: string
}

export function useCheckout() {
  const { post, get } = useApi()
  const config = useRuntimeConfig()
  const client = useSupabaseClient()

  const loading = ref(false)
  const error = ref('')
  const priceChanges = ref<PriceChange[]>([])

  let stripe: Stripe | null = null
  let elements: StripeElements | null = null

  async function getHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {}
    try {
      const { data: { session } } = await client.auth.getSession()
      if (session?.access_token) {
        headers['Authorization'] = `Bearer ${session.access_token}`
      }
    } catch {}
    if (import.meta.client) {
      const sessionId = localStorage.getItem('session_id')
      if (sessionId) {
        headers['X-Session-ID'] = sessionId
      }
    }
    return headers
  }

  async function startCheckout(email: string, shippingAddress: Record<string, string>) {
    loading.value = true
    error.value = ''
    priceChanges.value = []

    try {
      const headers = await getHeaders()
      const data = await post<StartCheckoutResponse>('/checkout/start', {
        email,
        shipping_address: shippingAddress,
      }, headers)
      return data
    } catch (err: any) {
      if (err.statusCode === 409 && err.data?.price_changes) {
        priceChanges.value = err.data.price_changes
        return null
      }
      error.value = err.data?.error || 'Checkout failed'
      return null
    } finally {
      loading.value = false
    }
  }

  async function initStripe(clientSecret: string) {
    stripe = await loadStripe(config.public.stripeKey as string)
    if (!stripe) throw new Error('Failed to load Stripe')
    elements = stripe.elements({ clientSecret })
    const paymentElement = elements.create('payment')
    paymentElement.mount('#payment-element')
  }

  async function confirmPayment(orderId: string) {
    if (!stripe || !elements) throw new Error('Stripe not initialized')
    const { error: stripeError } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: `${window.location.origin}/order/${orderId}`,
      },
    })
    // If confirmPayment returns, it means there was an error (success redirects)
    if (stripeError) {
      return stripeError.message || 'Payment failed'
    }
    return null
  }

  async function getOrder(orderId: string): Promise<OrderResponse> {
    const headers = await getHeaders()
    return await get<OrderResponse>(`/orders/${orderId}`, headers)
  }

  return { loading, error, priceChanges, startCheckout, initStripe, confirmPayment, getOrder }
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/nuxt.config.ts frontend/composables/useCheckout.ts
git commit -m "feat: add Stripe.js setup and useCheckout composable"
```

---

## Task 9: Frontend — Checkout Page

**Files:**
- Create: `frontend/pages/checkout.vue`

- [ ] **Step 1: Create the checkout page**

```vue
<!-- frontend/pages/checkout.vue -->
<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Checkout</h1>

    <!-- Loading cart -->
    <div v-if="cartLoading" class="text-gray-500">Loading...</div>

    <!-- Empty cart redirect -->
    <div v-else-if="!cart?.items?.length" class="text-center py-16">
      <p class="text-gray-500 mb-4">Your cart is empty</p>
      <NuxtLink to="/catalog" class="text-primary-600 hover:text-primary-700 font-medium">
        Browse catalog
      </NuxtLink>
    </div>

    <div v-else>
      <!-- Order Summary -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Order Summary</h2>
        <div v-for="item in cart.items" :key="item.id" class="flex justify-between py-2 border-b border-gray-100 last:border-0">
          <div>
            <span class="font-medium text-gray-900">{{ item.skus.products.name }}</span>
            <span class="text-sm text-gray-500 ml-2">{{ item.skus.sku_code }}</span>
            <span class="text-sm text-gray-500 ml-2">x{{ item.quantity }}</span>
          </div>
          <span class="text-gray-900">${{ (item.unit_price * item.quantity).toFixed(2) }}</span>
        </div>
        <div class="flex justify-between mt-4 pt-3 border-t border-gray-200 text-lg font-bold text-gray-900">
          <span>Total</span>
          <span>${{ cartTotal.toFixed(2) }}</span>
        </div>
      </div>

      <!-- Price Change Warning -->
      <div v-if="priceChanges.length" class="bg-yellow-50 border border-yellow-200 rounded-xl p-4 mb-6">
        <p class="font-medium text-yellow-800">Some prices have been updated:</p>
        <ul class="mt-2 text-sm text-yellow-700">
          <li v-for="change in priceChanges" :key="change.sku_id">
            {{ change.sku_code }}: ${{ change.old_price.toFixed(2) }} &rarr; ${{ change.new_price.toFixed(2) }}
          </li>
        </ul>
        <p class="mt-2 text-sm text-yellow-700">Your cart has been updated. Please review and try again.</p>
      </div>

      <!-- Step 1: Shipping Form -->
      <div v-if="step === 'shipping'" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Shipping Information</h2>
        <form @submit.prevent="handleStartCheckout" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input v-model="form.email" type="email" required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Full Name</label>
            <input v-model="form.name" type="text" required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
            <input v-model="form.line1" type="text" required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Apartment, suite, etc. (optional)</label>
            <input v-model="form.line2" type="text"
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">City</label>
              <input v-model="form.city" type="text" required
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">State / Province</label>
              <input v-model="form.state" type="text"
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">ZIP / Postal Code</label>
              <input v-model="form.zip" type="text" required
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Country</label>
              <input v-model="form.country" type="text" required
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
          </div>

          <button type="submit" :disabled="checkoutLoading"
            class="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed">
            {{ checkoutLoading ? 'Processing...' : 'Continue to Payment' }}
          </button>

          <p v-if="checkoutError" class="text-red-600 text-sm text-center">{{ checkoutError }}</p>
        </form>
      </div>

      <!-- Step 2: Payment -->
      <div v-if="step === 'payment'" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Payment</h2>
        <div id="payment-element" class="mb-6"></div>
        <button @click="handlePayment" :disabled="paying"
          class="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed">
          {{ paying ? 'Processing payment...' : `Pay $${cartTotal.toFixed(2)}` }}
        </button>
        <p v-if="paymentError" class="text-red-600 text-sm text-center mt-2">{{ paymentError }}</p>
        <button @click="step = 'shipping'" class="w-full text-sm text-gray-500 hover:text-gray-700 mt-3">
          Back to shipping
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { cart, loading: cartLoading, total: cartTotal, refresh: refreshCart } = useCart()
const { loading: checkoutLoading, error: checkoutError, priceChanges, startCheckout, initStripe, confirmPayment } = useCheckout()

const user = useSupabaseUser()
const step = ref<'shipping' | 'payment'>('shipping')
const orderId = ref('')
const paying = ref(false)
const paymentError = ref('')

const form = reactive({
  email: '',
  name: '',
  line1: '',
  line2: '',
  city: '',
  state: '',
  zip: '',
  country: 'US',
})

onMounted(async () => {
  await refreshCart()
  // Pre-fill email from auth if available
  if (user.value?.email) {
    form.email = user.value.email
  }
})

async function handleStartCheckout() {
  const result = await startCheckout(form.email, {
    name: form.name,
    line1: form.line1,
    line2: form.line2,
    city: form.city,
    state: form.state,
    zip: form.zip,
    country: form.country,
  })

  if (!result) {
    // Price change (409) or error — stay on shipping step, refresh cart
    if (priceChanges.value.length) {
      await refreshCart()
    }
    return
  }

  orderId.value = result.order_id
  step.value = 'payment'

  await nextTick()
  await initStripe(result.client_secret)
}

async function handlePayment() {
  paying.value = true
  paymentError.value = ''

  const errMsg = await confirmPayment(orderId.value)
  if (errMsg) {
    paymentError.value = errMsg
  }
  // On success, Stripe redirects to /order/:id — this code only runs on error
  paying.value = false
}
</script>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/pages/checkout.vue
git commit -m "feat: add checkout page with shipping form and Stripe Payment Element"
```

---

## Task 10: Frontend — Order Confirmation Page

**Files:**
- Create: `frontend/pages/order/[id].vue`

- [ ] **Step 1: Create the order confirmation page**

```vue
<!-- frontend/pages/order/[id].vue -->
<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-12">

    <!-- Payment Failed -->
    <div v-if="redirectStatus === 'failed'" class="text-center py-16">
      <div class="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <span class="text-red-600 text-3xl font-bold">&times;</span>
      </div>
      <h1 class="text-2xl font-bold text-gray-900 mb-2">Payment Failed</h1>
      <p class="text-gray-600 mb-6">Your payment could not be processed. Please try again.</p>
      <NuxtLink to="/cart" class="text-primary-600 hover:text-primary-700 font-medium">
        Return to Cart
      </NuxtLink>
    </div>

    <!-- Processing / Waiting for webhook -->
    <div v-else-if="!order || order.status === 'draft'" class="text-center py-16">
      <div class="w-10 h-10 border-4 border-primary-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
      <p class="text-gray-600">Confirming your payment...</p>
    </div>

    <!-- Order Confirmed -->
    <div v-else-if="order.status === 'paid'">
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <span class="text-green-600 text-3xl font-bold">&check;</span>
        </div>
        <h1 class="text-2xl font-bold text-gray-900 mb-1">Order Confirmed!</h1>
        <p class="text-gray-600">Thank you for your purchase.</p>
      </div>

      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div class="flex justify-between text-sm mb-3">
          <span class="text-gray-500">Order ID</span>
          <span class="font-mono text-gray-700">{{ order.id.slice(0, 8) }}...</span>
        </div>
        <div class="flex justify-between text-sm mb-4">
          <span class="text-gray-500">Confirmation sent to</span>
          <span class="text-gray-700">{{ order.email }}</span>
        </div>

        <div class="border-t border-gray-200 pt-4">
          <div v-for="item in order.items" :key="item.sku_code" class="flex justify-between py-2">
            <div>
              <span class="text-gray-900">{{ item.product_name }}</span>
              <span class="text-sm text-gray-500 ml-2">{{ item.sku_code }}</span>
              <span class="text-sm text-gray-500 ml-2">&times;{{ item.quantity }}</span>
            </div>
            <span class="text-gray-900">${{ (item.unit_price * item.quantity).toFixed(2) }}</span>
          </div>
        </div>

        <div class="border-t border-gray-200 pt-4 mt-2 flex justify-between font-bold text-gray-900">
          <span>Total</span>
          <span>${{ order.total.toFixed(2) }}</span>
        </div>
      </div>

      <div class="text-center mt-8">
        <NuxtLink to="/catalog" class="text-primary-600 hover:text-primary-700 font-medium">
          Continue Shopping
        </NuxtLink>
      </div>
    </div>

    <!-- Cancelled -->
    <div v-else-if="order.status === 'cancelled'" class="text-center py-16">
      <h1 class="text-2xl font-bold text-gray-900 mb-2">Order Cancelled</h1>
      <p class="text-gray-600 mb-6">This order has been cancelled.</p>
      <NuxtLink to="/cart" class="text-primary-600 hover:text-primary-700 font-medium">
        Return to Cart
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { OrderResponse } from '~/composables/useCheckout'

const route = useRoute()
const { getOrder } = useCheckout()

const orderId = route.params.id as string
const redirectStatus = (route.query.redirect_status as string) || ''

const order = ref<OrderResponse | null>(null)

onMounted(() => {
  if (redirectStatus === 'failed') return
  pollOrderStatus()
})

async function pollOrderStatus() {
  for (let i = 0; i < 15; i++) {
    try {
      order.value = await getOrder(orderId)
      if (order.value?.status === 'paid' || order.value?.status === 'cancelled') return
    } catch {}
    await new Promise(resolve => setTimeout(resolve, 2000))
  }
}
</script>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/pages/order/[id].vue
git commit -m "feat: add order confirmation page with status polling"
```

---

## Task 11: Frontend — Wire Cart Checkout Button

**Files:**
- Modify: `frontend/pages/cart.vue`

- [ ] **Step 1: Enable the checkout button**

In `frontend/pages/cart.vue`, replace the disabled checkout button:

```html
<button
  disabled
  class="mt-4 w-full bg-gray-300 text-white py-3 rounded-lg font-medium cursor-not-allowed"
  title="Checkout coming in Plan 3"
>
  Proceed to Checkout
</button>
```

With a working link:

```html
<NuxtLink
  to="/checkout"
  class="mt-4 w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors text-center block"
>
  Proceed to Checkout
</NuxtLink>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/pages/cart.vue
git commit -m "feat: wire Proceed to Checkout button to checkout page"
```

---

## Local Development Notes

**Stripe test keys:** Add to your shell environment or `.env`:

```
STRIPE_SECRET_KEY=sk_test_...        # Go API
STRIPE_WEBHOOK_SECRET=whsec_...      # Go API (from Stripe CLI)
NUXT_PUBLIC_STRIPE_KEY=pk_test_...   # Frontend
```

**Webhook forwarding for local dev:** Run the Stripe CLI to forward webhooks to your local API:

```
stripe listen --forward-to localhost:9090/stripe/webhook
```

This prints the webhook signing secret (`whsec_...`) — use that as `STRIPE_WEBHOOK_SECRET`.
