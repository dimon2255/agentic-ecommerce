package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func TestOrderHandler_List(t *testing.T) {
	orders := []OrderSummary{
		{ID: "order-1", Status: "paid", Email: "a@b.com", Total: 99.99, CreatedAt: time.Now()},
		{ID: "order-2", Status: "shipped", Email: "c@d.com", Total: 49.99, CreatedAt: time.Now()},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Range", "0-1/2")
		json.NewEncoder(w).Encode(orders)
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	repo := NewSupabaseRepository(db)
	audit := NewAuditService(db)
	handler := NewOrderHandler(repo, audit)

	req := httptest.NewRequest("GET", "/?page=1&per_page=20", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp pagination.Response[OrderSummary]
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Items) != 2 {
		t.Errorf("expected 2 orders, got %d", len(resp.Items))
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
}

func TestOrderHandler_UpdateStatus(t *testing.T) {
	callLog := map[string]int{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			callLog["get"]++
			// Return order for GetOrder
			w.Header().Set("Content-Range", "0-0/1")
			json.NewEncoder(w).Encode([]OrderDetail{
				{ID: "order-1", Status: "paid", Email: "a@b.com"},
			})
		case "PATCH":
			callLog["patch"]++
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		case "POST":
			callLog["post"]++
			// Audit log insert
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("[]"))
		}
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	repo := NewSupabaseRepository(db)
	audit := NewAuditService(db)
	handler := NewOrderHandler(repo, audit)

	body, _ := json.Marshal(UpdateOrderStatusRequest{Status: "shipped"})
	req := httptest.NewRequest("PATCH", "/order-1/status", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "admin-1")
	req = req.WithContext(ctx)

	// Set chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "order-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.UpdateStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "shipped" {
		t.Errorf("expected status shipped, got %s", resp["status"])
	}
}

func TestOrderHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	handler := NewOrderHandler(nil, nil)

	body, _ := json.Marshal(UpdateOrderStatusRequest{Status: "invalid"})
	req := httptest.NewRequest("PATCH", "/order-1/status", bytes.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "order-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.UpdateStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
