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
