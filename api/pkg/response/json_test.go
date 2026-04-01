package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
	"github.com/dimon2255/agentic-ecommerce/api/internal/requestid"
)

// helper to create a request with request ID in context
func requestWithID(id string) *http.Request {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Request-ID", id)
	rec := httptest.NewRecorder()
	var captured *http.Request
	requestid.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r
	})).ServeHTTP(rec, r)
	return captured
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	Success(w, http.StatusOK, map[string]string{"name": "test"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("unexpected Content-Type: %s", ct)
	}

	var body struct {
		Data map[string]string `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if body.Data["name"] != "test" {
		t.Errorf("expected data.name=test, got %s", body.Data["name"])
	}
}

func TestErrorFromAppError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := requestWithID("req-123")
	err := apperror.NewNotFound("product")

	ErrorFromAppError(w, r, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var body errorEnvelope
	json.NewDecoder(w.Body).Decode(&body)
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", body.Error.Code)
	}
	if body.Error.Message != "product not found" {
		t.Errorf("expected 'product not found', got %s", body.Error.Message)
	}
	if body.Error.RequestID != "req-123" {
		t.Errorf("expected req-123, got %s", body.Error.RequestID)
	}
}

func TestErrorFromAppError_InvalidInput(t *testing.T) {
	w := httptest.NewRecorder()
	r := requestWithID("req-456")
	err := apperror.NewInvalidInput("validation failed", map[string]string{
		"name": "required",
		"slug": "too short",
	})

	ErrorFromAppError(w, r, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var body errorEnvelope
	json.NewDecoder(w.Body).Decode(&body)
	if body.Error.Code != "INVALID_INPUT" {
		t.Errorf("expected INVALID_INPUT, got %s", body.Error.Code)
	}
	if len(body.Error.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(body.Error.Fields))
	}
	if body.Error.Fields["name"] != "required" {
		t.Errorf("expected fields.name=required, got %s", body.Error.Fields["name"])
	}
}

func TestErrorFromAppError_Conflict(t *testing.T) {
	w := httptest.NewRecorder()
	r := requestWithID("req-789")
	err := apperror.NewConflict("price changed", []string{"sku-1", "sku-2"})

	ErrorFromAppError(w, r, err)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
			Data []string `json:"data"`
		} `json:"error"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if body.Error.Code != "CONFLICT" {
		t.Errorf("expected CONFLICT, got %s", body.Error.Code)
	}
	if len(body.Error.Data) != 2 {
		t.Errorf("expected 2 data items, got %d", len(body.Error.Data))
	}
}

func TestErrorFromAppError_Internal(t *testing.T) {
	w := httptest.NewRecorder()
	r := requestWithID("req-internal")
	err := apperror.NewInternal("server error", fmt.Errorf("db failed"))

	ErrorFromAppError(w, r, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var body errorEnvelope
	json.NewDecoder(w.Body).Decode(&body)
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %s", body.Error.Code)
	}
	// Message should NOT contain "db failed"
	if body.Error.Message != "server error" {
		t.Errorf("expected 'server error', got %s", body.Error.Message)
	}
}

func TestErrorFromAppError_GenericError(t *testing.T) {
	w := httptest.NewRecorder()
	r := requestWithID("req-generic")
	err := fmt.Errorf("something unexpected")

	ErrorFromAppError(w, r, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var body errorEnvelope
	json.NewDecoder(w.Body).Decode(&body)
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %s", body.Error.Code)
	}
	if body.Error.Message != "internal server error" {
		t.Errorf("expected 'internal server error', got %s", body.Error.Message)
	}
}

func TestLegacyJSON_StillWorks(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, map[string]string{"name": "test"})

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["name"] != "test" {
		t.Errorf("expected name=test, got %s", body["name"])
	}
}

func TestLegacyError_StillWorks(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad input")

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "bad input" {
		t.Errorf("expected error='bad input', got %s", body["error"])
	}
}
