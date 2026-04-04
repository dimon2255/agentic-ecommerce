package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssistantRateLimiter_AuthenticatedUser(t *testing.T) {
	rl := NewAssistantRateLimiter(20, 3, 5, 2)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 3 requests should pass (minute burst limit = 3)
	for i := range 3 {
		req := httptest.NewRequest("POST", "/assistant/stream", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, "user-123")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// 4th request should be rejected (minute burst)
	req := httptest.NewRequest("POST", "/assistant/stream", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "user-123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("4th request: expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on 429")
	}
}

func TestAssistantRateLimiter_GuestWithSessionID(t *testing.T) {
	rl := NewAssistantRateLimiter(20, 5, 5, 2)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 2 guest requests should pass (guest minute burst = 2)
	for i := range 2 {
		req := httptest.NewRequest("POST", "/assistant/stream", nil)
		req.Header.Set("X-Session-ID", "sess-abc")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// 3rd guest request should be rejected
	req := httptest.NewRequest("POST", "/assistant/stream", nil)
	req.Header.Set("X-Session-ID", "sess-abc")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("3rd guest request: expected 429, got %d", rr.Code)
	}
}

func TestAssistantRateLimiter_NoIdentity(t *testing.T) {
	rl := NewAssistantRateLimiter(20, 5, 5, 2)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/assistant/stream", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAssistantRateLimiter_SeparateUsersIndependent(t *testing.T) {
	rl := NewAssistantRateLimiter(20, 2, 5, 2)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// User A uses 2 requests (exhausts burst)
	for range 2 {
		req := httptest.NewRequest("POST", "/assistant/stream", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, "user-A")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatal("user-A request should pass")
		}
	}

	// User B should still have full quota
	req := httptest.NewRequest("POST", "/assistant/stream", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "user-B")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("user-B request: expected 200, got %d", rr.Code)
	}
}
