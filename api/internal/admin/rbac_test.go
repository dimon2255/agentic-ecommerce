package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func TestRBACService_GetPermissions(t *testing.T) {
	expectedPerms := []string{"catalog:read", "catalog:write", "orders:read"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/v1/rpc/get_user_permissions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", 404)
			return
		}

		var params map[string]string
		json.NewDecoder(r.Body).Decode(&params)
		if params["p_user_id"] != "user-123" {
			t.Errorf("expected user-123, got %s", params["p_user_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedPerms)
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	svc := NewRBACService(db, 5*time.Minute)

	perms, err := svc.GetPermissions(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 3 {
		t.Fatalf("expected 3 permissions, got %d", len(perms))
	}
	if perms[0] != "catalog:read" {
		t.Errorf("expected catalog:read, got %s", perms[0])
	}
}

func TestRBACService_CachesPermissions(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"catalog:read"})
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	svc := NewRBACService(db, 5*time.Minute)

	// First call — hits server
	_, err := svc.GetPermissions(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call — should use cache
	_, err = svc.GetPermissions(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 RPC call (cached), got %d", callCount)
	}
}

func TestRBACService_InvalidateCache(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"catalog:read"})
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	svc := NewRBACService(db, 5*time.Minute)

	svc.GetPermissions(context.Background(), "user-123")
	svc.InvalidateCache("user-123")
	svc.GetPermissions(context.Background(), "user-123")

	if callCount != 2 {
		t.Errorf("expected 2 RPC calls (cache invalidated), got %d", callCount)
	}
}

func TestRBACService_EmptyPermissions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	svc := NewRBACService(db, 5*time.Minute)

	perms, err := svc.GetPermissions(context.Background(), "customer-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 0 {
		t.Errorf("expected 0 permissions, got %d", len(perms))
	}
}

func TestRBACService_GetUserRoles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/rest/v1/user_roles":
			json.NewEncoder(w).Encode([]map[string]string{
				{"role_id": "role-1"},
				{"role_id": "role-2"},
			})
		case r.URL.Path == "/rest/v1/roles":
			json.NewEncoder(w).Encode([]map[string]string{
				{"name": "admin"},
				{"name": "catalog_manager"},
			})
		default:
			http.Error(w, "not found", 404)
		}
	}))
	defer ts.Close()

	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	svc := NewRBACService(db, 5*time.Minute)

	roles, err := svc.GetUserRoles(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[0] != "admin" {
		t.Errorf("expected admin, got %s", roles[0])
	}
}
