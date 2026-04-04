package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
	"github.com/dimon2255/agentic-ecommerce/api/internal/requestid"
)

// JSON writes data directly as JSON. Used by existing handlers during migration.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Error writes a flat error response. Used by existing handlers during migration.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

// --- New structured envelope functions (Phase 2+ handlers use these) ---

// Success writes a structured success envelope: {"data": ...}
func Success(w http.ResponseWriter, status int, data any) {
	JSON(w, status, envelope{Data: data})
}

// ErrorFromAppError writes a structured error from a typed AppError.
// Falls back to 500 INTERNAL_ERROR for non-AppError errors.
func ErrorFromAppError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := requestid.Get(r.Context())

	appErr, ok := apperror.As(err)
	if !ok {
		slog.Error("unhandled error", "request_id", requestID, "error", err)
		JSON(w, http.StatusInternalServerError, errorEnvelope{
			Error: errorBody{
				Code:      "INTERNAL_ERROR",
				Message:   "internal server error",
				RequestID: requestID,
			},
		})
		return
	}

	// Log internal errors server-side with full detail
	if ie, ok := err.(*apperror.InternalError); ok {
		slog.Error("internal error", "request_id", requestID, "error", ie.Error())
	}

	body := errorEnvelope{
		Error: errorBody{
			Code:      appErr.Code(),
			Message:   appErr.Message(),
			RequestID: requestID,
		},
	}

	// Attach field-level validation errors
	if ie, ok := err.(*apperror.InvalidInputError); ok && len(ie.Fields) > 0 {
		body.Error.Fields = ie.Fields
	}

	// Attach conflict data (e.g., price changes)
	if ce, ok := err.(*apperror.ConflictError); ok && ce.Data != nil {
		body.Error.Data = ce.Data
	}

	JSON(w, appErr.HTTPStatus(), body)
}

type envelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	Data      any               `json:"data,omitempty"`
}
