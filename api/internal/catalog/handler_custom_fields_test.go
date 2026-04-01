package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestCustomFieldHandler(supabaseHandler http.HandlerFunc) (*CustomFieldHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key", 10*time.Second)
	repo := NewSupabaseRepository(client)
	svc := NewService(repo)
	handler := NewCustomFieldHandler(svc)
	return handler, server
}

func TestListCustomFields(t *testing.T) {
	handler, server := setupTestCustomFieldHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]CustomField{
			{ID: "1", EntityType: "product", EntityID: "prod-1", Key: "supplier", Value: "Acme"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/custom-fields?entity_type=product&entity_id=prod-1", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []CustomField
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 field, got %d", len(result))
	}
}

func TestCreateCustomField(t *testing.T) {
	handler, server := setupTestCustomFieldHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]CustomField{{ID: "1", EntityType: "product", EntityID: "prod-1", Key: "supplier", Value: "Acme"}})
	})
	defer server.Close()

	body := `{"key":"supplier","value":"Acme"}`
	req := httptest.NewRequest("POST", "/custom-fields?entity_type=product&entity_id=prod-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
