package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type CategoryHandler struct {
	svc Service
}

func NewCategoryHandler(svc Service) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{slug}", h.GetBySlug)
	r.Patch("/{slug}", h.Update)
	r.Delete("/{slug}", h.DeleteBySlug)
	return r
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := CategoryFilter{Params: pagination.ParseFromQuery(r)}
	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		filter.ParentID = &parentID
	}

	categories, total, err := h.svc.ListCategories(r.Context(), filter)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, pagination.NewResponse(categories, total, filter.Params))
}

func (h *CategoryHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	cat, err := h.svc.GetCategoryBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, cat)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, cat)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.UpdateCategory(r.Context(), chi.URLParam(r, "slug"), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, cat)
}

func (h *CategoryHandler) DeleteBySlug(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteCategory(r.Context(), chi.URLParam(r, "slug")); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
