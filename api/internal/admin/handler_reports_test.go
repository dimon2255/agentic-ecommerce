package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func TestReportsHandler_Dashboard(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DashboardKPIs{
			{TotalOrders: 42, TotalRevenue: 1234.56, ActiveProducts: 13, TotalCustomers: 8},
		})
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	repo := NewSupabaseRepository(db)
	handler := NewReportsHandler(repo)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	rec := httptest.NewRecorder()

	handler.Dashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var kpis DashboardKPIs
	json.NewDecoder(rec.Body).Decode(&kpis)

	if kpis.TotalOrders != 42 {
		t.Errorf("expected 42 orders, got %d", kpis.TotalOrders)
	}
	if kpis.TotalRevenue != 1234.56 {
		t.Errorf("expected revenue 1234.56, got %f", kpis.TotalRevenue)
	}
}

func TestReportsHandler_SalesByDay(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SalesByDay{
			{Day: "2026-04-03", OrderCount: 5, Revenue: 499.95},
			{Day: "2026-04-02", OrderCount: 3, Revenue: 299.97},
		})
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	repo := NewSupabaseRepository(db)
	handler := NewReportsHandler(repo)

	req := httptest.NewRequest("GET", "/sales?date_from=2026-04-01&date_to=2026-04-04", nil)
	rec := httptest.NewRecorder()

	handler.SalesByDay(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var sales []SalesByDay
	json.NewDecoder(rec.Body).Decode(&sales)

	if len(sales) != 2 {
		t.Errorf("expected 2 days, got %d", len(sales))
	}
}

func TestReportsHandler_TokenUsage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]TokenUsageByDay{
			{Day: "2026-04-03", InputTokens: 10000, OutputTokens: 5000, RequestCount: 50},
		})
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	repo := NewSupabaseRepository(db)
	handler := NewReportsHandler(repo)

	req := httptest.NewRequest("GET", "/token-usage", nil)
	rec := httptest.NewRecorder()

	handler.TokenUsageByDay(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var usage []TokenUsageByDay
	json.NewDecoder(rec.Body).Decode(&usage)

	if len(usage) != 1 {
		t.Errorf("expected 1 day, got %d", len(usage))
	}
	if usage[0].InputTokens != 10000 {
		t.Errorf("expected 10000 input tokens, got %d", usage[0].InputTokens)
	}
}
