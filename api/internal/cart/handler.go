package cart

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type CartHandler struct {
	svc Service
}

func NewCartHandler(svc Service) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetCart)
	r.Post("/items", h.AddItem)
	r.Patch("/items/{itemId}", h.UpdateItem)
	r.Delete("/items/{itemId}", h.RemoveItem)
	r.Post("/merge", h.MergeCart)
	return r
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")

	resp, err := h.svc.GetCart(r.Context(), userID, sessionID)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.AddItem(r.Context(), userID, sessionID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")
	itemID := chi.URLParam(r, "itemId")

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.UpdateItem(r.Context(), userID, sessionID, itemID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")
	itemID := chi.URLParam(r, "itemId")

	if err := h.svc.RemoveItem(r.Context(), userID, sessionID, itemID); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) MergeCart(w http.ResponseWriter, r *http.Request) {
	userID, hasUser := middleware.GetUserID(r.Context())
	if !hasUser || userID == "" {
		response.ErrorFromAppError(w, r, apperror.NewUnauthorized("authentication required"))
		return
	}

	var req MergeCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.MergeCart(r.Context(), userID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
