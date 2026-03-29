package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CategoryHandler struct {
	db *supabase.Client
}

func NewCategoryHandler(db *supabase.Client) *CategoryHandler {
	return &CategoryHandler{db: db}
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
	parentID := r.URL.Query().Get("parent_id")

	query := h.db.From("categories").Select("*").Order("name", "asc")
	if parentID == "null" {
		query = query.Is("parent_id", "null")
	} else if parentID != "" {
		query = query.Eq("parent_id", parentID)
	}

	var categories []Category
	if err := query.Execute(&categories); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	if categories == nil {
		categories = []Category{}
	}
	response.JSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var category Category
	err := h.db.From("categories").Select("*").Eq("slug", slug).Single().Execute(&category)
	if err != nil {
		response.Error(w, http.StatusNotFound, "category not found")
		return
	}

	response.JSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" {
		response.Error(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	var created []Category
	if err := h.db.From("categories").Insert(req).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var updated []Category
	err := h.db.From("categories").Eq("slug", slug).Update(req).Execute(&updated)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update category")
		return
	}

	if len(updated) == 0 {
		response.Error(w, http.StatusNotFound, "category not found")
		return
	}

	response.JSON(w, http.StatusOK, updated[0])
}

func (h *CategoryHandler) DeleteBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	err := h.db.From("categories").Eq("slug", slug).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
