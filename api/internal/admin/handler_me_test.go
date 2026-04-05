package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func TestMeHandler_GetMe(t *testing.T) {
	// Mock Supabase that returns permissions and roles
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/rest/v1/rpc/get_user_permissions":
			json.NewEncoder(w).Encode([]string{"catalog:read", "orders:read"})
		case r.URL.Path == "/rest/v1/user_roles":
			json.NewEncoder(w).Encode([]map[string]string{{"role_id": "role-1"}})
		case r.URL.Path == "/rest/v1/roles":
			json.NewEncoder(w).Encode([]map[string]string{{"name": "admin"}})
		}
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	rbac := NewRBACService(db, 5*time.Minute)
	handler := NewMeHandler(rbac)

	req := httptest.NewRequest("GET", "/", nil)
	// Set user ID and permissions in context (as middleware would)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	ctx = context.WithValue(ctx, middleware.PermissionsKey, []string{"catalog:read", "orders:read"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["user_id"] != "user-123" {
		t.Errorf("expected user_id user-123, got %v", resp["user_id"])
	}
	perms, ok := resp["permissions"].([]any)
	if !ok || len(perms) != 2 {
		t.Errorf("expected 2 permissions, got %v", resp["permissions"])
	}
	roles, ok := resp["roles"].([]any)
	if !ok || len(roles) != 1 {
		t.Errorf("expected 1 role, got %v", resp["roles"])
	}
}
