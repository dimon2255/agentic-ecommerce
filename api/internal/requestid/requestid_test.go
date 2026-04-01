package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_GeneratesWhenMissing(t *testing.T) {
	var capturedID string
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = Get(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected request ID to be generated")
	}
	if rec.Header().Get(Header) == "" {
		t.Error("expected X-Request-ID response header")
	}
	if rec.Header().Get(Header) != capturedID {
		t.Error("response header should match context value")
	}
}

func TestMiddleware_UsesExistingHeader(t *testing.T) {
	var capturedID string
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = Get(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(Header, "my-custom-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID != "my-custom-id" {
		t.Errorf("expected my-custom-id, got %s", capturedID)
	}
	if rec.Header().Get(Header) != "my-custom-id" {
		t.Errorf("expected header my-custom-id, got %s", rec.Header().Get(Header))
	}
}

func TestMiddleware_UniquePerRequest(t *testing.T) {
	ids := make(map[string]bool)
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids[Get(r.Context())] = true
		w.WriteHeader(http.StatusOK)
	}))

	for range 100 {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if len(ids) != 100 {
		t.Errorf("expected 100 unique IDs, got %d", len(ids))
	}
}

func TestGet_EmptyContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	id := Get(req.Context())
	if id != "" {
		t.Errorf("expected empty string, got %s", id)
	}
}
