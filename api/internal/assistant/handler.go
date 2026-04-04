package assistant

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// Handler handles HTTP requests for the assistant domain.
type Handler struct {
	svc Service
}

// NewHandler creates an assistant handler backed by the given service.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes returns a chi.Router with assistant routes that use the standard request timeout.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Chat)
	r.Post("/tools", h.ChatWithTools)
	return r
}

// StreamRoute returns the SSE streaming handler (mounted separately to bypass timeout middleware).
func (h *Handler) StreamRoute() http.HandlerFunc {
	return h.StreamChat
}

// Chat handles POST / — processes a user message and returns an AI response.
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Chat(r.Context(), userID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ChatWithTools handles POST /tools — processes a message using Claude tool use.
func (h *Handler) ChatWithTools(w http.ResponseWriter, r *http.Request) {
	userID, isGuest := resolveIdentity(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "authentication or session ID required")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.ChatWithTools(r.Context(), userID, isGuest, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// StreamChat handles POST /stream — streams an AI response via SSE with tool use.
func (h *Handler) StreamChat(w http.ResponseWriter, r *http.Request) {
	userID, isGuest := resolveIdentity(r)
	if userID == "" {
		http.Error(w, `{"error":"authentication or session ID required"}`, http.StatusUnauthorized)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Use context with 2-minute timeout for the streaming operation
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	emit := func(event, data string) {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
		flusher.Flush()
	}

	if err := h.svc.StreamChat(ctx, userID, isGuest, req, emit); err != nil {
		slog.ErrorContext(ctx, "StreamChat error", "error", err)
		b, _ := json.Marshal(map[string]string{"message": "An error occurred while processing your request."})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(b))
		flusher.Flush()
	}
}

// resolveIdentity extracts user identity from auth context or X-Session-ID header.
// Returns (userID, isGuest). For guests, userID is a deterministic UUID derived from session ID.
func resolveIdentity(r *http.Request) (string, bool) {
	if userID, ok := middleware.GetUserID(r.Context()); ok {
		return userID, false
	}
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		return "", false
	}
	return guestUserID(sessionID), true
}

// guestUserID generates a deterministic UUID-like string from a session ID.
// This avoids schema changes to chat_sessions.user_id while keeping guest sessions trackable.
func guestUserID(sessionID string) string {
	h := sha256.Sum256([]byte("guest:" + sessionID))
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		h[0:4], h[4:6], h[6:8], h[8:10], h[10:16])
}
