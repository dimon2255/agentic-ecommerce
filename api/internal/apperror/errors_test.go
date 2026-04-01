package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestNotFoundError(t *testing.T) {
	err := NewNotFound("product")
	if err.Code() != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", err.Code())
	}
	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("expected 404, got %d", err.HTTPStatus())
	}
	if err.Message() != "product not found" {
		t.Errorf("expected 'product not found', got %s", err.Message())
	}
}

func TestInvalidInputError(t *testing.T) {
	fields := map[string]string{"name": "required", "slug": "too short"}
	err := NewInvalidInput("validation failed", fields)
	if err.Code() != "INVALID_INPUT" {
		t.Errorf("expected INVALID_INPUT, got %s", err.Code())
	}
	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.HTTPStatus())
	}
	if len(err.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(err.Fields))
	}
}

func TestConflictError(t *testing.T) {
	err := NewConflict("price changed", map[string]any{"old": 10, "new": 15})
	if err.Code() != "CONFLICT" {
		t.Errorf("expected CONFLICT, got %s", err.Code())
	}
	if err.HTTPStatus() != http.StatusConflict {
		t.Errorf("expected 409, got %d", err.HTTPStatus())
	}
	if err.Data == nil {
		t.Error("expected data to be set")
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := NewUnauthorized("missing token")
	if err.Code() != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %s", err.Code())
	}
	if err.HTTPStatus() != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", err.HTTPStatus())
	}
}

func TestForbiddenError(t *testing.T) {
	err := NewForbidden("access denied")
	if err.Code() != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got %s", err.Code())
	}
	if err.HTTPStatus() != http.StatusForbidden {
		t.Errorf("expected 403, got %d", err.HTTPStatus())
	}
}

func TestInternalError(t *testing.T) {
	underlying := fmt.Errorf("db connection failed")
	err := NewInternal("server error", underlying)
	if err.Code() != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %s", err.Code())
	}
	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", err.HTTPStatus())
	}
	// Message should NOT expose underlying error
	if err.Message() != "server error" {
		t.Errorf("expected 'server error', got %s", err.Message())
	}
	// Error() includes underlying for logging
	if err.Error() != "server error: db connection failed" {
		t.Errorf("unexpected Error(): %s", err.Error())
	}
	// Unwrap exposes underlying
	if !errors.Is(err, underlying) {
		t.Error("expected Unwrap to return underlying error")
	}
}

func TestInternalError_NilUnderlying(t *testing.T) {
	err := NewInternal("something broke", nil)
	if err.Error() != "something broke" {
		t.Errorf("unexpected Error(): %s", err.Error())
	}
}

func TestAs(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", NewNotFound("category"))
	appErr, ok := As(err)
	if !ok {
		t.Fatal("expected As to find AppError")
	}
	if appErr.Code() != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", appErr.Code())
	}
}

func TestAs_NonAppError(t *testing.T) {
	err := fmt.Errorf("plain error")
	_, ok := As(err)
	if ok {
		t.Error("expected As to return false for non-AppError")
	}
}
