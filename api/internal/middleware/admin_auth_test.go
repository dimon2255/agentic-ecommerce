package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockPermissionResolver implements PermissionResolver for tests.
type mockPermissionResolver struct {
	perms map[string][]string // userID -> permissions
	err   error
}

func (m *mockPermissionResolver) GetPermissions(_ context.Context, userID string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.perms[userID], nil
}

func TestRequirePermission_AllGranted(t *testing.T) {
	resolver := &mockPermissionResolver{
		perms: map[string][]string{
			"admin-1": {"catalog:read", "catalog:write", "orders:read"},
		},
	}

	handler := RequirePermission(resolver, "catalog:read", "catalog:write")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "admin-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequirePermission_MissingOne(t *testing.T) {
	resolver := &mockPermissionResolver{
		perms: map[string][]string{
			"admin-1": {"catalog:read"},
		},
	}

	handler := RequirePermission(resolver, "catalog:read", "catalog:write")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "admin-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequirePermission_NoAuth(t *testing.T) {
	resolver := &mockPermissionResolver{}

	handler := RequirePermission(resolver, "catalog:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}),
	)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequirePermission_ResolverError(t *testing.T) {
	resolver := &mockPermissionResolver{err: fmt.Errorf("db down")}

	handler := RequirePermission(resolver, "catalog:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "admin-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRequireAnyPermission_HasOne(t *testing.T) {
	resolver := &mockPermissionResolver{
		perms: map[string][]string{
			"admin-1": {"orders:read"},
		},
	}

	handler := RequireAnyPermission(resolver, "catalog:read", "orders:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "admin-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAnyPermission_HasNone(t *testing.T) {
	resolver := &mockPermissionResolver{
		perms: map[string][]string{
			"admin-1": {"settings:read"},
		},
	}

	handler := RequireAnyPermission(resolver, "catalog:read", "orders:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "admin-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequirePermission_PermissionsCachedInContext(t *testing.T) {
	resolver := &mockPermissionResolver{
		perms: map[string][]string{
			"admin-1": {"catalog:read", "catalog:write"},
		},
	}

	var capturedPerms []string
	handler := RequirePermission(resolver, "catalog:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPerms = GetPermissions(r.Context())
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "admin-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if len(capturedPerms) != 2 {
		t.Errorf("expected 2 permissions in context, got %d", len(capturedPerms))
	}
}

func TestRequirePermission_ReusesExistingContextPermissions(t *testing.T) {
	callCount := 0
	resolver := &mockPermissionResolver{
		perms: map[string][]string{
			"admin-1": {"catalog:read", "orders:read"},
		},
	}
	// Wrap resolver to count calls
	countingResolver := &countingPermissionResolver{inner: resolver, count: &callCount}

	// Chain two permission middlewares
	inner := RequirePermission(countingResolver, "orders:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	outer := RequirePermission(countingResolver, "catalog:read")(inner)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "admin-1"))
	rec := httptest.NewRecorder()

	outer.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// Should only resolve once, second middleware reuses context
	if callCount != 1 {
		t.Errorf("expected resolver called once, got %d", callCount)
	}
}

type countingPermissionResolver struct {
	inner PermissionResolver
	count *int
}

func (c *countingPermissionResolver) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	*c.count++
	return c.inner.GetPermissions(ctx, userID)
}
