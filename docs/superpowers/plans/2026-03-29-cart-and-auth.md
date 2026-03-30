# Cart & Auth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Supabase Auth (registration, login, JWT validation) and a server-side shopping cart with guest support and cart merge on login.

**Architecture:** Nuxt 3 handles auth UI via `@nuxtjs/supabase` module (client-side Supabase Auth). Go API validates JWTs from Supabase for authenticated endpoints. Cart is server-side — Go API manages cart state via Supabase PostgREST. Guests use an `X-Session-ID` header; on login, guest carts merge into the user's cart.

**Tech Stack:** Nuxt 3, Vue 3, Tailwind CSS, @nuxtjs/supabase v2, Go 1.22+, chi router, golang-jwt/v5, Supabase (PostgREST, Auth), PostgreSQL

---

## File Map

### Go API (`api/`)

| File | Responsibility |
|------|---------------|
| `api/internal/middleware/auth.go` | JWT validation middleware (OptionalAuth, RequireAuth) + context helpers |
| `api/internal/middleware/auth_test.go` | Auth middleware tests |
| `api/internal/cart/models.go` | Cart and cart item domain types, request/response structs |
| `api/internal/cart/handler.go` | Cart CRUD handlers (get, add, update, remove, merge) |
| `api/internal/cart/handler_test.go` | Cart handler tests |
| `api/cmd/server/main.go` | Updated: cart routes, auth middleware, CORS header |

### Nuxt Frontend (`frontend/`)

| File | Responsibility |
|------|---------------|
| `frontend/composables/useApi.ts` | Updated: accepts optional headers parameter |
| `frontend/composables/useCart.ts` | Cart state management + API calls with auth/session headers |
| `frontend/pages/auth/login.vue` | Login page (SPA) |
| `frontend/pages/auth/register.vue` | Registration page (SPA) |
| `frontend/pages/cart.vue` | Cart page with item list, quantity controls, totals (SPA) |
| `frontend/components/CartItem.vue` | Single cart item row with quantity +/- and remove |
| `frontend/layouts/default.vue` | Updated: auth state in header, cart badge |
| `frontend/pages/product/[slug].vue` | Updated: working "Add to Cart" button |
| `frontend/nuxt.config.ts` | Updated: auth route rules |

### Supabase (`supabase/`)

| File | Responsibility |
|------|---------------|
| `supabase/migrations/00003_carts.sql` | Carts and cart_items tables with RLS |

---

## Task 1: Database Migration — Carts Schema

**Files:**
- Create: `supabase/migrations/00003_carts.sql`

- [ ] **Step 1: Create carts migration**

Create `supabase/migrations/00003_carts.sql`:

```sql
-- Carts and cart items

CREATE TYPE cart_status AS ENUM ('active', 'merged', 'expired');

-- Carts table
CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    session_id TEXT NOT NULL,
    status cart_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_carts_user_id ON carts(user_id);
CREATE INDEX idx_carts_session_id ON carts(session_id);
CREATE INDEX idx_carts_status ON carts(status);

-- Cart items table
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    sku_id UUID NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_cart_items_sku_id ON cart_items(sku_id);
CREATE UNIQUE INDEX idx_cart_items_cart_sku ON cart_items(cart_id, sku_id);

-- Reuse existing trigger function from migration 00001
CREATE TRIGGER update_carts_updated_at
    BEFORE UPDATE ON carts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_cart_items_updated_at
    BEFORE UPDATE ON cart_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Row Level Security
ALTER TABLE carts ENABLE ROW LEVEL SECURITY;
ALTER TABLE cart_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own carts" ON carts
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Service role full access on carts" ON carts
    FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "Users can view own cart items" ON cart_items
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM carts WHERE carts.id = cart_items.cart_id AND carts.user_id = auth.uid())
    );

CREATE POLICY "Service role full access on cart_items" ON cart_items
    FOR ALL USING (auth.role() = 'service_role');
```

- [ ] **Step 2: Apply migration**

```bash
cd supabase
supabase db reset
```

Expected: Migration applies without errors. Tables `carts` and `cart_items` exist.

- [ ] **Step 3: Verify tables exist**

Open Supabase Studio at `http://127.0.0.1:54323` and confirm `carts` and `cart_items` tables are present with correct columns.

- [ ] **Step 4: Commit**

```bash
git add supabase/migrations/00003_carts.sql
git commit -m "feat: add carts and cart_items schema migration"
```

---

## Task 2: Go API — Auth Middleware

**Files:**
- Create: `api/internal/middleware/auth.go`
- Create: `api/internal/middleware/auth_test.go`

- [ ] **Step 1: Install JWT dependency**

```bash
cd api
go get github.com/golang-jwt/jwt/v5
```

- [ ] **Step 2: Write auth middleware tests**

Create `api/internal/middleware/auth_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-jwt-secret-key-at-least-32-chars!!"

func generateTestToken(t *testing.T, secret string, userID string, exp time.Time) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"exp":  exp.Unix(),
		"aud":  "authenticated",
		"role": "authenticated",
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenStr
}

func TestOptionalAuth_WithValidToken(t *testing.T) {
	auth := NewAuthMiddleware(testJWTSecret)

	var capturedUserID string
	handler := auth.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := GetUserID(r.Context())
		if ok {
			capturedUserID = id
		}
		w.WriteHeader(http.StatusOK)
	}))

	token := generateTestToken(t, testJWTSecret, "user-123", time.Now().Add(time.Hour))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedUserID != "user-123" {
		t.Errorf("expected user-123, got %s", capturedUserID)
	}
}

func TestOptionalAuth_WithoutToken(t *testing.T) {
	auth := NewAuthMiddleware(testJWTSecret)

	var hasUserID bool
	handler := auth.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if hasUserID {
		t.Error("expected no user ID in context")
	}
}

func TestOptionalAuth_WithInvalidToken(t *testing.T) {
	auth := NewAuthMiddleware(testJWTSecret)

	var hasUserID bool
	handler := auth.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if hasUserID {
		t.Error("expected no user ID with invalid token")
	}
}

func TestRequireAuth_WithValidToken(t *testing.T) {
	auth := NewAuthMiddleware(testJWTSecret)

	var capturedUserID string
	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := GetUserID(r.Context())
		capturedUserID = id
		w.WriteHeader(http.StatusOK)
	}))

	token := generateTestToken(t, testJWTSecret, "user-456", time.Now().Add(time.Hour))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedUserID != "user-456" {
		t.Errorf("expected user-456, got %s", capturedUserID)
	}
}

func TestRequireAuth_WithoutToken(t *testing.T) {
	auth := NewAuthMiddleware(testJWTSecret)

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_WithExpiredToken(t *testing.T) {
	auth := NewAuthMiddleware(testJWTSecret)

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	token := generateTestToken(t, testJWTSecret, "user-789", time.Now().Add(-time.Hour))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
cd api
go test ./internal/middleware/ -v
```

Expected: Compilation errors — `NewAuthMiddleware`, `GetUserID` not defined.

- [ ] **Step 4: Implement auth middleware**

Create `api/internal/middleware/auth.go`:

```go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: []byte(jwtSecret)}
}

// OptionalAuth extracts user ID from JWT if present. Request proceeds regardless.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr != "" {
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return m.jwtSecret, nil
			})
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if sub, ok := claims["sub"].(string); ok {
						ctx := context.WithValue(r.Context(), UserIDKey, sub)
						r = r.WithContext(ctx)
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth rejects requests without a valid JWT. Returns 401 on failure.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr == "" {
			response.Error(w, http.StatusUnauthorized, "missing authorization token")
			return
		}
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			response.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "invalid token claims")
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "invalid token subject")
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, sub)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// GetUserID extracts the authenticated user ID from the request context.
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}
```

- [ ] **Step 5: Run tests**

```bash
cd api
go test ./internal/middleware/ -v
```

Expected: All 6 tests pass.

- [ ] **Step 6: Commit**

```bash
git add api/internal/middleware/auth.go api/internal/middleware/auth_test.go api/go.mod api/go.sum
git commit -m "feat: add JWT auth middleware with optional and required modes"
```

---

## Task 3: Go API — Cart Models

**Files:**
- Create: `api/internal/cart/models.go`

- [ ] **Step 1: Define cart domain types**

Create `api/internal/cart/models.go`:

```go
package cart

import "time"

// --- Database Models ---

type Cart struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id"`
	SessionID string    `json:"session_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartItem struct {
	ID        string    `json:"id"`
	CartID    string    `json:"cart_id"`
	SKUID     string    `json:"sku_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartItemWithSKU includes nested SKU/product data from PostgREST embedded select.
// PostgREST uses table names as JSON keys for embedded resources.
type CartItemWithSKU struct {
	ID        string    `json:"id"`
	CartID    string    `json:"cart_id"`
	SKUID     string    `json:"sku_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SKU       SKUEmbed  `json:"skus"`
}

type SKUEmbed struct {
	SKUCode       string       `json:"sku_code"`
	PriceOverride *float64     `json:"price_override"`
	Product       ProductEmbed `json:"products"`
}

type ProductEmbed struct {
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	BasePrice float64  `json:"base_price"`
	Images    []string `json:"images"`
}

// SKUForPrice is used when looking up current SKU price for cart snapshot.
type SKUForPrice struct {
	PriceOverride *float64     `json:"price_override"`
	Product       ProductEmbed `json:"products"`
}

// --- Request/Response Types ---

type CartResponse struct {
	ID    string            `json:"id"`
	Items []CartItemWithSKU `json:"items"`
}

type AddItemRequest struct {
	SKUID    string `json:"sku_id"`
	Quantity int    `json:"quantity"`
}

type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

type MergeCartRequest struct {
	SessionID string `json:"session_id"`
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd api
go build ./internal/cart/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add api/internal/cart/models.go
git commit -m "feat: add cart domain models and request types"
```

---

## Task 4: Go API — Cart Handlers (Get & Add Item)

**Files:**
- Create: `api/internal/cart/handler.go`
- Create: `api/internal/cart/handler_test.go`

- [ ] **Step 1: Write tests for GetCart and AddItem**

Create `api/internal/cart/handler_test.go`:

```go
package cart

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestCartHandler(supabaseHandler http.HandlerFunc) (*CartHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd api
go test ./internal/cart/ -v
```

Expected: Compilation errors — `CartHandler`, `NewCartHandler`, `GetCart`, `AddItem` not defined.

- [ ] **Step 3: Implement cart handler (GetCart, AddItem, helpers)**

Create `api/internal/cart/handler.go`:

```go
package cart

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CartHandler struct {
	db *supabase.Client
}

func NewCartHandler(db *supabase.Client) *CartHandler {
	return &CartHandler{db: db}
}

func (h *CartHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetCart)
	r.Post("/items", h.AddItem)
	r.Patch("/items/{itemId}", h.UpdateItem)
	r.Delete("/items/{itemId}", h.RemoveItem)
	r.Post("/merge", h.MergeCart)
	return r
}

// findActiveCart looks up the active cart for the current user or session.
func (h *CartHandler) findActiveCart(r *http.Request) *Cart {
	userID, hasUser := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")

	var carts []Cart
	q := h.db.From("carts").Select("*").Eq("status", "active")
	if hasUser && userID != "" {
		q = q.Eq("user_id", userID)
	} else if sessionID != "" {
		q = q.Eq("session_id", sessionID).Is("user_id", "null")
	} else {
		return nil
	}

	if err := q.Execute(&carts); err != nil || len(carts) == 0 {
		return nil
	}
	return &carts[0]
}

// findOrCreateCart returns the active cart or creates one.
func (h *CartHandler) findOrCreateCart(r *http.Request) (*Cart, error) {
	if cart := h.findActiveCart(r); cart != nil {
		return cart, nil
	}

	userID, _ := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		return nil, fmt.Errorf("session ID required")
	}

	newCart := map[string]any{
		"session_id": sessionID,
		"status":     "active",
	}
	if userID != "" {
		newCart["user_id"] = userID
	}

	var created []Cart
	if err := h.db.From("carts").Insert(newCart).Execute(&created); err != nil {
		return nil, fmt.Errorf("create cart: %w", err)
	}
	if len(created) == 0 {
		return nil, fmt.Errorf("cart not returned after creation")
	}
	return &created[0], nil
}

// getCartResponse fetches the full cart with enriched items.
func (h *CartHandler) getCartResponse(cartID string) (*CartResponse, error) {
	var items []CartItemWithSKU
	err := h.db.From("cart_items").
		Select("*,skus(sku_code,price_override,products(name,slug,base_price,images))").
		Eq("cart_id", cartID).
		Execute(&items)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []CartItemWithSKU{}
	}
	return &CartResponse{ID: cartID, Items: items}, nil
}

// lookupSKUPrice fetches the current price for a SKU.
func (h *CartHandler) lookupSKUPrice(skuID string) (float64, error) {
	var skus []SKUForPrice
	err := h.db.From("skus").
		Select("price_override,products(base_price)").
		Eq("id", skuID).
		Execute(&skus)
	if err != nil {
		return 0, err
	}
	if len(skus) == 0 {
		return 0, fmt.Errorf("SKU not found")
	}
	sku := skus[0]
	if sku.PriceOverride != nil {
		return *sku.PriceOverride, nil
	}
	return sku.Product.BasePrice, nil
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, hasUser := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")
	if (!hasUser || userID == "") && sessionID == "" {
		response.Error(w, http.StatusBadRequest, "authentication or session ID required")
		return
	}

	cart := h.findActiveCart(r)
	if cart == nil {
		response.JSON(w, http.StatusOK, CartResponse{Items: []CartItemWithSKU{}})
		return
	}

	resp, err := h.getCartResponse(cart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch cart items")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SKUID == "" || req.Quantity < 1 {
		response.Error(w, http.StatusBadRequest, "sku_id is required and quantity must be at least 1")
		return
	}

	cart, err := h.findOrCreateCart(r)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get or create cart")
		return
	}

	unitPrice, err := h.lookupSKUPrice(req.SKUID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid SKU")
		return
	}

	// Check for existing item with same SKU
	var existing []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("cart_id", cart.ID).Eq("sku_id", req.SKUID).
		Execute(&existing); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to check existing items")
		return
	}

	if len(existing) > 0 {
		// Increment quantity
		newQty := existing[0].Quantity + req.Quantity
		var updated []CartItem
		if err := h.db.From("cart_items").
			Update(map[string]any{"quantity": newQty}).
			Eq("id", existing[0].ID).
			Execute(&updated); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to update item quantity")
			return
		}
	} else {
		// Insert new item
		var inserted []CartItem
		if err := h.db.From("cart_items").Insert(map[string]any{
			"cart_id":    cart.ID,
			"sku_id":     req.SKUID,
			"quantity":   req.Quantity,
			"unit_price": unitPrice,
		}).Execute(&inserted); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to add item to cart")
			return
		}
	}

	resp, err := h.getCartResponse(cart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch updated cart")
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not implemented")
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not implemented")
}

func (h *CartHandler) MergeCart(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not implemented")
}
```

- [ ] **Step 4: Run tests**

```bash
cd api
go test ./internal/cart/ -v
```

Expected: All 4 tests pass (GetCart empty, GetCart with items, requires session/auth, add item creates cart, validation errors).

- [ ] **Step 5: Commit**

```bash
git add api/internal/cart/handler.go api/internal/cart/handler_test.go
git commit -m "feat: add cart handlers for get cart and add item (TDD)"
```

---

## Task 5: Go API — Cart Handlers (Update, Remove, Merge)

**Files:**
- Modify: `api/internal/cart/handler_test.go`
- Modify: `api/internal/cart/handler.go`

- [ ] **Step 1: Write tests for UpdateItem, RemoveItem, MergeCart**

Append to `api/internal/cart/handler_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify new tests fail**

```bash
cd api
go test ./internal/cart/ -v -run "TestUpdateItem|TestRemoveItem|TestMergeCart"
```

Expected: Tests fail — `UpdateItem`, `RemoveItem`, `MergeCart` return 501 Not Implemented.

- [ ] **Step 3: Implement UpdateItem, RemoveItem, MergeCart**

Replace the stub methods in `api/internal/cart/handler.go`:

```go
func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Quantity < 1 {
		response.Error(w, http.StatusBadRequest, "quantity must be at least 1")
		return
	}

	cart := h.findActiveCart(r)
	if cart == nil {
		response.Error(w, http.StatusNotFound, "cart not found")
		return
	}

	// Verify item belongs to this cart
	var items []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("id", itemID).Eq("cart_id", cart.ID).
		Execute(&items); err != nil || len(items) == 0 {
		response.Error(w, http.StatusNotFound, "cart item not found")
		return
	}

	var updated []CartItem
	if err := h.db.From("cart_items").
		Update(map[string]any{"quantity": req.Quantity}).
		Eq("id", itemID).
		Execute(&updated); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update item")
		return
	}

	resp, err := h.getCartResponse(cart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch updated cart")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")

	cart := h.findActiveCart(r)
	if cart == nil {
		response.Error(w, http.StatusNotFound, "cart not found")
		return
	}

	// Verify item belongs to this cart
	var items []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("id", itemID).Eq("cart_id", cart.ID).
		Execute(&items); err != nil || len(items) == 0 {
		response.Error(w, http.StatusNotFound, "cart item not found")
		return
	}

	if err := h.db.From("cart_items").Delete().
		Eq("id", itemID).Execute(nil); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to remove item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) MergeCart(w http.ResponseWriter, r *http.Request) {
	userID, hasUser := middleware.GetUserID(r.Context())
	if !hasUser || userID == "" {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req MergeCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SessionID == "" {
		response.Error(w, http.StatusBadRequest, "session_id is required")
		return
	}

	// Find guest cart
	var guestCarts []Cart
	if err := h.db.From("carts").Select("*").
		Eq("session_id", req.SessionID).
		Is("user_id", "null").
		Eq("status", "active").
		Execute(&guestCarts); err != nil || len(guestCarts) == 0 {
		// No guest cart to merge — return current user cart
		userCart := h.findUserCart(userID)
		if userCart == nil {
			response.JSON(w, http.StatusOK, CartResponse{Items: []CartItemWithSKU{}})
			return
		}
		resp, err := h.getCartResponse(userCart.ID)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to fetch cart")
			return
		}
		response.JSON(w, http.StatusOK, resp)
		return
	}
	guestCart := guestCarts[0]

	// Find or create user cart
	userCart := h.findUserCart(userID)
	if userCart == nil {
		var created []Cart
		if err := h.db.From("carts").Insert(map[string]any{
			"user_id":    userID,
			"session_id": req.SessionID,
			"status":     "active",
		}).Execute(&created); err != nil || len(created) == 0 {
			response.Error(w, http.StatusInternalServerError, "failed to create user cart")
			return
		}
		userCart = &created[0]
	}

	// Fetch guest cart items
	var guestItems []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("cart_id", guestCart.ID).
		Execute(&guestItems); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch guest cart items")
		return
	}

	// Move each guest item to user cart
	for _, item := range guestItems {
		// Check for duplicate SKU in user cart
		var existing []CartItem
		h.db.From("cart_items").Select("*").
			Eq("cart_id", userCart.ID).Eq("sku_id", item.SKUID).
			Execute(&existing)

		if len(existing) > 0 {
			// Increment quantity on existing user cart item
			newQty := existing[0].Quantity + item.Quantity
			h.db.From("cart_items").
				Update(map[string]any{"quantity": newQty}).
				Eq("id", existing[0].ID).
				Execute(nil)
			// Delete guest item
			h.db.From("cart_items").Delete().
				Eq("id", item.ID).Execute(nil)
		} else {
			// Move item to user cart
			h.db.From("cart_items").
				Update(map[string]any{"cart_id": userCart.ID}).
				Eq("id", item.ID).
				Execute(nil)
		}
	}

	// Mark guest cart as merged
	h.db.From("carts").
		Update(map[string]any{"status": "merged"}).
		Eq("id", guestCart.ID).
		Execute(nil)

	resp, err := h.getCartResponse(userCart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch merged cart")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// findUserCart looks up the active cart for a specific user ID.
func (h *CartHandler) findUserCart(userID string) *Cart {
	var carts []Cart
	if err := h.db.From("carts").Select("*").
		Eq("user_id", userID).
		Eq("status", "active").
		Execute(&carts); err != nil || len(carts) == 0 {
		return nil
	}
	return &carts[0]
}
```

- [ ] **Step 4: Run all cart tests**

```bash
cd api
go test ./internal/cart/ -v
```

Expected: All 8 tests pass.

- [ ] **Step 5: Commit**

```bash
git add api/internal/cart/handler.go api/internal/cart/handler_test.go
git commit -m "feat: add cart update, remove, and merge handlers (TDD)"
```

---

## Task 6: Go API — Wire Up Router

**Files:**
- Modify: `api/cmd/server/main.go`

- [ ] **Step 1: Wire cart handlers and auth middleware into the router**

Update `api/cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = "http://127.0.0.1:54321"
	}
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseKey == "" {
		log.Fatal("SUPABASE_SERVICE_ROLE_KEY is required")
	}
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("SUPABASE_JWT_SECRET is required")
	}

	db := supabase.NewClient(supabaseURL, supabaseKey)
	auth := middleware.NewAuthMiddleware(jwtSecret)

	categoryHandler := catalog.NewCategoryHandler(db)
	attributeHandler := catalog.NewAttributeHandler(db)
	productHandler := catalog.NewProductHandler(db)
	skuHandler := catalog.NewSKUHandler(db)
	customFieldHandler := catalog.NewCustomFieldHandler(db)
	cartHandler := cart.NewCartHandler(db)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Session-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/categories", categoryHandler.Routes())
		r.Route("/categories/{categoryId}/attributes", func(r chi.Router) {
			r.Mount("/", attributeHandler.Routes())
		})
		r.Mount("/products", productHandler.Routes())
		r.Route("/products/{productId}/skus", func(r chi.Router) {
			r.Mount("/", skuHandler.Routes())
		})
		r.Mount("/custom-fields", customFieldHandler.Routes())

		// Cart routes — OptionalAuth so both guests and users can access
		r.Group(func(r chi.Router) {
			r.Use(auth.OptionalAuth)
			r.Mount("/cart", cartHandler.Routes())
		})
	})

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("API server listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd api
go build ./cmd/server/
```

Expected: No errors.

- [ ] **Step 3: Run all tests**

```bash
cd api
go test ./... -v
```

Expected: All tests pass (middleware + cart + catalog).

- [ ] **Step 4: Commit**

```bash
git add api/cmd/server/main.go
git commit -m "feat: wire cart routes with auth middleware into API router"
```

---

## Task 7: Nuxt — Update API Composable & Config

**Files:**
- Modify: `frontend/composables/useApi.ts`
- Modify: `frontend/nuxt.config.ts`

- [ ] **Step 1: Update useApi to accept optional headers**

Update `frontend/composables/useApi.ts`:

```ts
export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase

  async function get<T>(path: string, headers?: Record<string, string>): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, { headers })
  }

  async function post<T>(path: string, body: any, headers?: Record<string, string>): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'POST',
      body,
      headers,
    })
  }

  async function patch<T>(path: string, body: any, headers?: Record<string, string>): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'PATCH',
      body,
      headers,
    })
  }

  async function del(path: string, headers?: Record<string, string>): Promise<void> {
    await $fetch(`${baseURL}/api/v1${path}`, {
      method: 'DELETE',
      headers,
    })
  }

  return { get, post, patch, del }
}
```

- [ ] **Step 2: Add auth route rules to nuxt.config.ts**

Add the auth route rule to `frontend/nuxt.config.ts` `routeRules`:

```ts
'/auth/**': { ssr: false },
```

The full `routeRules` block should be:

```ts
routeRules: {
  '/': { ssr: true },
  '/catalog/**': { ssr: true },
  '/product/**': { ssr: true },
  '/cart': { ssr: false },
  '/checkout': { ssr: false },
  '/auth/**': { ssr: false },
  '/account/**': { ssr: false },
},
```

- [ ] **Step 3: Commit**

```bash
git add frontend/composables/useApi.ts frontend/nuxt.config.ts
git commit -m "feat: update useApi with optional headers, add auth route rules"
```

---

## Task 8: Nuxt — Auth Pages

**Files:**
- Create: `frontend/pages/auth/login.vue`
- Create: `frontend/pages/auth/register.vue`

- [ ] **Step 1: Create login page**

Create `frontend/pages/auth/login.vue`:

```vue
<template>
  <div class="min-h-[60vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-bold text-gray-900 text-center mb-8">Sign in</h1>

      <div v-if="errorMsg" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
        {{ errorMsg }}
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            autocomplete="email"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            autocomplete="current-password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-primary-600 text-white py-2.5 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          {{ loading ? 'Signing in...' : 'Sign in' }}
        </button>
      </form>

      <p class="mt-6 text-center text-sm text-gray-500">
        Don't have an account?
        <NuxtLink to="/auth/register" class="text-primary-600 hover:text-primary-700 font-medium">
          Create one
        </NuxtLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const client = useSupabaseClient()
const router = useRouter()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  loading.value = true
  errorMsg.value = ''

  const { error } = await client.auth.signInWithPassword({
    email: email.value,
    password: password.value,
  })

  if (error) {
    errorMsg.value = error.message
    loading.value = false
    return
  }

  // Merge guest cart if session exists
  try {
    const sessionId = localStorage.getItem('session_id')
    if (sessionId) {
      const { post } = useApi()
      const { data: { session } } = await client.auth.getSession()
      if (session?.access_token) {
        await post('/cart/merge', { session_id: sessionId }, {
          'Authorization': `Bearer ${session.access_token}`,
          'X-Session-ID': sessionId,
        })
      }
    }
  } catch {
    // Cart merge is best-effort
  }

  router.push('/')
}
</script>
```

- [ ] **Step 2: Create register page**

Create `frontend/pages/auth/register.vue`:

```vue
<template>
  <div class="min-h-[60vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-bold text-gray-900 text-center mb-8">Create account</h1>

      <div v-if="errorMsg" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
        {{ errorMsg }}
      </div>
      <div v-if="successMsg" class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700">
        {{ successMsg }}
      </div>

      <form @submit.prevent="handleRegister" class="space-y-4">
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            autocomplete="email"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            minlength="6"
            autocomplete="new-password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <div>
          <label for="confirm" class="block text-sm font-medium text-gray-700 mb-1">Confirm password</label>
          <input
            id="confirm"
            v-model="confirmPassword"
            type="password"
            required
            minlength="6"
            autocomplete="new-password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-primary-600 text-white py-2.5 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          {{ loading ? 'Creating account...' : 'Create account' }}
        </button>
      </form>

      <p class="mt-6 text-center text-sm text-gray-500">
        Already have an account?
        <NuxtLink to="/auth/login" class="text-primary-600 hover:text-primary-700 font-medium">
          Sign in
        </NuxtLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const client = useSupabaseClient()
const router = useRouter()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

async function handleRegister() {
  loading.value = true
  errorMsg.value = ''
  successMsg.value = ''

  if (password.value !== confirmPassword.value) {
    errorMsg.value = 'Passwords do not match'
    loading.value = false
    return
  }

  const { error } = await client.auth.signUp({
    email: email.value,
    password: password.value,
  })

  if (error) {
    errorMsg.value = error.message
    loading.value = false
    return
  }

  // In local dev, Supabase auto-confirms. In prod, email confirmation may be required.
  successMsg.value = 'Account created! Redirecting...'
  setTimeout(() => router.push('/auth/login'), 1500)
}
</script>
```

- [ ] **Step 3: Verify pages load**

```bash
cd frontend
npm run dev
```

Navigate to `http://localhost:3000/auth/login` and `http://localhost:3000/auth/register`. Both should render the forms without errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/pages/auth/login.vue frontend/pages/auth/register.vue
git commit -m "feat: add login and register pages with Supabase Auth"
```

---

## Task 9: Nuxt — Cart Composable

**Files:**
- Create: `frontend/composables/useCart.ts`

- [ ] **Step 1: Create cart composable**

Create `frontend/composables/useCart.ts`:

```ts
interface CartItemSKU {
  sku_code: string
  price_override: number | null
  products: {
    name: string
    slug: string
    base_price: number
    images: string[]
  }
}

export interface CartItem {
  id: string
  sku_id: string
  quantity: number
  unit_price: number
  skus: CartItemSKU
}

export interface CartData {
  id: string
  items: CartItem[]
}

export function useCart() {
  const { get, post, patch, del } = useApi()
  const client = useSupabaseClient()

  const cart = useState<CartData | null>('cart', () => null)
  const loading = useState('cart-loading', () => false)

  async function getHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {}

    try {
      const { data: { session } } = await client.auth.getSession()
      if (session?.access_token) {
        headers['Authorization'] = `Bearer ${session.access_token}`
      }
    } catch {
      // No auth available
    }

    if (import.meta.client) {
      let sessionId = localStorage.getItem('session_id')
      if (!sessionId) {
        sessionId = crypto.randomUUID()
        localStorage.setItem('session_id', sessionId)
      }
      headers['X-Session-ID'] = sessionId
    }

    return headers
  }

  async function refresh() {
    if (!import.meta.client) return
    loading.value = true
    try {
      const headers = await getHeaders()
      cart.value = await get<CartData>('/cart', headers)
    } catch {
      cart.value = null
    } finally {
      loading.value = false
    }
  }

  async function addItem(skuId: string, quantity: number = 1) {
    const headers = await getHeaders()
    cart.value = await post<CartData>('/cart/items', { sku_id: skuId, quantity }, headers)
  }

  async function updateItem(itemId: string, quantity: number) {
    const headers = await getHeaders()
    cart.value = await patch<CartData>(`/cart/items/${itemId}`, { quantity }, headers)
  }

  async function removeItem(itemId: string) {
    const headers = await getHeaders()
    await del(`/cart/items/${itemId}`, headers)
    await refresh()
  }

  const itemCount = computed(() => {
    if (!cart.value?.items) return 0
    return cart.value.items.reduce((sum, item) => sum + item.quantity, 0)
  })

  const total = computed(() => {
    if (!cart.value?.items) return 0
    return cart.value.items.reduce((sum, item) => sum + item.unit_price * item.quantity, 0)
  })

  return { cart, loading, itemCount, total, refresh, addItem, updateItem, removeItem }
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd frontend
npx nuxi typecheck
```

Expected: No type errors related to useCart.

- [ ] **Step 3: Commit**

```bash
git add frontend/composables/useCart.ts
git commit -m "feat: add cart composable with auth headers and session tracking"
```

---

## Task 10: Nuxt — Cart Page

**Files:**
- Create: `frontend/components/CartItem.vue`
- Create: `frontend/pages/cart.vue`

- [ ] **Step 1: Create CartItem component**

Create `frontend/components/CartItem.vue`:

```vue
<template>
  <div class="flex items-center gap-4 py-4 border-b border-gray-100">
    <div class="w-16 h-16 bg-gray-100 rounded-lg flex-shrink-0 overflow-hidden">
      <img
        v-if="item.skus.products.images?.length"
        :src="item.skus.products.images[0]"
        :alt="item.skus.products.name"
        class="w-full h-full object-cover"
      />
    </div>
    <div class="flex-1 min-w-0">
      <NuxtLink :to="`/product/${item.skus.products.slug}`" class="text-sm font-medium text-gray-900 hover:text-primary-600 truncate block">
        {{ item.skus.products.name }}
      </NuxtLink>
      <p class="text-xs text-gray-500 mt-0.5">{{ item.skus.sku_code }}</p>
      <p class="text-sm font-medium text-gray-900 mt-1">${{ item.unit_price.toFixed(2) }}</p>
    </div>
    <div class="flex items-center gap-2">
      <button
        @click="$emit('update', item.id, item.quantity - 1)"
        :disabled="item.quantity <= 1 || updating"
        class="w-8 h-8 flex items-center justify-center rounded-md border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        -
      </button>
      <span class="w-8 text-center text-sm font-medium">{{ item.quantity }}</span>
      <button
        @click="$emit('update', item.id, item.quantity + 1)"
        :disabled="updating"
        class="w-8 h-8 flex items-center justify-center rounded-md border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-40"
      >
        +
      </button>
    </div>
    <div class="text-right w-20">
      <p class="text-sm font-medium text-gray-900">${{ (item.unit_price * item.quantity).toFixed(2) }}</p>
    </div>
    <button
      @click="$emit('remove', item.id)"
      :disabled="updating"
      class="text-gray-400 hover:text-red-500 transition-colors disabled:opacity-40"
      title="Remove item"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import type { CartItem as CartItemType } from '~/composables/useCart'

defineProps<{
  item: CartItemType
  updating: boolean
}>()

defineEmits<{
  update: [itemId: string, quantity: number]
  remove: [itemId: string]
}>()
</script>
```

- [ ] **Step 2: Create cart page**

Create `frontend/pages/cart.vue`:

```vue
<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Shopping Cart</h1>

    <div v-if="loading" class="text-gray-500">Loading cart...</div>

    <div v-else-if="!cart?.items?.length" class="text-center py-16">
      <p class="text-gray-500 mb-4">Your cart is empty</p>
      <NuxtLink to="/catalog" class="text-primary-600 hover:text-primary-700 font-medium">
        Browse catalog
      </NuxtLink>
    </div>

    <div v-else>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 px-6">
        <CartItem
          v-for="item in cart.items"
          :key="item.id"
          :item="item"
          :updating="updating"
          @update="handleUpdate"
          @remove="handleRemove"
        />
      </div>

      <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div class="flex justify-between items-center text-lg font-bold text-gray-900">
          <span>Total</span>
          <span>${{ total.toFixed(2) }}</span>
        </div>
        <button
          disabled
          class="mt-4 w-full bg-gray-300 text-white py-3 rounded-lg font-medium cursor-not-allowed"
          title="Checkout coming in Plan 3"
        >
          Proceed to Checkout
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { cart, loading, total, refresh, updateItem, removeItem } = useCart()
const updating = ref(false)

onMounted(() => {
  refresh()
})

async function handleUpdate(itemId: string, quantity: number) {
  updating.value = true
  try {
    await updateItem(itemId, quantity)
  } finally {
    updating.value = false
  }
}

async function handleRemove(itemId: string) {
  updating.value = true
  try {
    await removeItem(itemId)
  } finally {
    updating.value = false
  }
}
</script>
```

- [ ] **Step 3: Verify cart page renders**

```bash
cd frontend
npm run dev
```

Navigate to `http://localhost:3000/cart`. Should show "Your cart is empty" with a link to the catalog.

- [ ] **Step 4: Commit**

```bash
git add frontend/components/CartItem.vue frontend/pages/cart.vue
git commit -m "feat: add cart page with item list, quantity controls, and totals"
```

---

## Task 11: Nuxt — Layout Update & Add to Cart

**Files:**
- Modify: `frontend/layouts/default.vue`
- Modify: `frontend/pages/product/[slug].vue`

- [ ] **Step 1: Update layout with auth state and cart badge**

Update `frontend/layouts/default.vue`:

```vue
<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <NuxtLink to="/" class="text-xl font-bold text-gray-900 tracking-tight">
            FlexShop
          </NuxtLink>
          <nav class="flex items-center gap-6">
            <NuxtLink to="/catalog" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors">
              Catalog
            </NuxtLink>
            <NuxtLink to="/cart" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors relative">
              Cart
              <ClientOnly>
                <span
                  v-if="itemCount > 0"
                  class="absolute -top-2 -right-4 bg-primary-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full"
                >
                  {{ itemCount > 9 ? '9+' : itemCount }}
                </span>
              </ClientOnly>
            </NuxtLink>
            <ClientOnly>
              <template v-if="user">
                <span class="text-sm text-gray-500">{{ user.email }}</span>
                <button
                  @click="handleLogout"
                  class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
                >
                  Sign out
                </button>
              </template>
              <template v-else>
                <NuxtLink to="/auth/login" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors">
                  Sign in
                </NuxtLink>
              </template>
            </ClientOnly>
          </nav>
        </div>
      </div>
    </header>
    <main>
      <slot />
    </main>
    <footer class="bg-white border-t border-gray-200 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <p class="text-sm text-gray-400 text-center">
          &copy; {{ new Date().getFullYear() }} FlexShop. All rights reserved.
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
const user = useSupabaseUser()
const client = useSupabaseClient()
const router = useRouter()
const { itemCount, refresh } = useCart()

onMounted(() => {
  refresh()
})

async function handleLogout() {
  await client.auth.signOut()
  router.push('/')
}
</script>
```

- [ ] **Step 2: Wire Add to Cart button on product page**

Update `frontend/pages/product/[slug].vue` — replace the `<button>` and add the `addToCart` handler:

Replace the existing button element:

```vue
<button
  class="mt-8 w-full bg-primary-600 text-white py-3 px-6 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
  :disabled="!selectedSku || addingToCart"
  @click="addToCart"
>
  {{ addingToCart ? 'Adding...' : selectedSku ? 'Add to Cart' : 'Select options' }}
</button>
<p v-if="addedMsg" class="mt-2 text-sm text-green-600 text-center">{{ addedMsg }}</p>
```

Add to the `<script setup>` section (after `selectedSku` ref):

```ts
const { addItem } = useCart()
const addingToCart = ref(false)
const addedMsg = ref('')

async function addToCart() {
  if (!selectedSku.value) return
  addingToCart.value = true
  addedMsg.value = ''
  try {
    await addItem(selectedSku.value.id)
    addedMsg.value = 'Added to cart!'
    setTimeout(() => { addedMsg.value = '' }, 2000)
  } catch {
    addedMsg.value = 'Failed to add to cart'
  } finally {
    addingToCart.value = false
  }
}
```

- [ ] **Step 3: Verify full flow**

```bash
cd frontend
npm run dev
```

1. Navigate to a product page
2. Select SKU options
3. Click "Add to Cart" — should show "Added to cart!" and the cart badge should increment
4. Navigate to `/cart` — should see the item with quantity controls
5. Click +/- to change quantity
6. Click X to remove item

- [ ] **Step 4: Commit**

```bash
git add frontend/layouts/default.vue frontend/pages/product/\[slug\].vue
git commit -m "feat: add auth state to layout, wire Add to Cart on product page"
```

---

## Task 12: Final Integration & Cleanup

**Files:**
- Modify: `PROGRESS.md`

- [ ] **Step 1: Run all Go tests**

```bash
cd api
go test ./... -v
```

Expected: All tests pass.

- [ ] **Step 2: Run full stack manually**

Start all services:

```bash
# Terminal 1: Supabase
cd supabase && supabase start

# Terminal 2: Go API
cd api && SUPABASE_SERVICE_ROLE_KEY=<key> SUPABASE_JWT_SECRET=<jwt-secret> go run cmd/server/main.go

# Terminal 3: Nuxt
cd frontend && npm run dev
```

Manual test checklist:
- [ ] Browse catalog pages (SSR still works)
- [ ] Register a new account at `/auth/register`
- [ ] Login at `/auth/login`
- [ ] Header shows email and "Sign out" button
- [ ] Add item to cart from product detail page
- [ ] Cart badge updates in header
- [ ] Cart page shows items with correct prices
- [ ] Update quantity with +/- buttons
- [ ] Remove item from cart
- [ ] Sign out — header shows "Sign in" link
- [ ] Guest cart works (add items without logging in)
- [ ] Login after guest cart — items merge into user cart

- [ ] **Step 3: Update PROGRESS.md**

Update `PROGRESS.md` to mark Plan 2 milestones:

```
- [x] Plan 2: Cart & Auth — implementation plan written
- [x] Plan 2: Cart & Auth — implemented
```

Update the `Current Phase` to `Plan 3 — Checkout & Payments` and the Plan 2 status to `Complete`.

- [ ] **Step 4: Commit and push**

```bash
git add -A
git commit -m "feat: complete Plan 2 — Cart & Auth"
git push origin HEAD
```
