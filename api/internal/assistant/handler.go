package assistant

import (
	"encoding/json"
	"net/http"

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

// Routes returns a chi.Router with the assistant routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Chat)
	r.Post("/tools", h.ChatWithTools)
	return r
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

	resp, err := h.svc.ChatWithTools(r.Context(), userID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
