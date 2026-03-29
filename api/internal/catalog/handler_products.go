package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type ProductHandler struct {
	db *supabase.Client
}

func NewProductHandler(db *supabase.Client) *ProductHandler {
	return &ProductHandler{db: db}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{slug}", h.GetBySlug)
	r.Patch("/{slug}", h.Update)
	r.Delete("/{slug}", h.DeleteBySlug)
	return r
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("category_id")

	query := h.db.From("products").Select("*").Order("created_at", "desc")
	if categoryID != "" {
		query = query.Eq("category_id", categoryID)
	}

	var products []Product
	if err := query.Execute(&products); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch products")
		return
	}

	if products == nil {
		products = []Product{}
	}
	response.JSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var product Product
	err := h.db.From("products").Select("*").Eq("slug", slug).Single().Execute(&product)
	if err != nil {
		response.Error(w, http.StatusNotFound, "product not found")
		return
	}

	response.JSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" || req.CategoryID == "" {
		response.Error(w, http.StatusBadRequest, "name, slug, and category_id are required")
		return
	}

	if req.Status == "" {
		req.Status = "draft"
	}

	var created []Product
	if err := h.db.From("products").Insert(req).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var updated []Product
	err := h.db.From("products").Eq("slug", slug).Update(req).Execute(&updated)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update product")
		return
	}

	if len(updated) == 0 {
		response.Error(w, http.StatusNotFound, "product not found")
		return
	}

	response.JSON(w, http.StatusOK, updated[0])
}

func (h *ProductHandler) DeleteBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	err := h.db.From("products").Eq("slug", slug).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
