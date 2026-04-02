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
	auth := NewAuthMiddleware(testJWTSecret, "", "", "")

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
	auth := NewAuthMiddleware(testJWTSecret, "", "", "")

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
	auth := NewAuthMiddleware(testJWTSecret, "", "", "")

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
	auth := NewAuthMiddleware(testJWTSecret, "", "", "")

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
	auth := NewAuthMiddleware(testJWTSecret, "", "", "")

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
	auth := NewAuthMiddleware(testJWTSecret, "", "", "")

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
