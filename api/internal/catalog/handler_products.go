package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type ProductHandler struct {
	svc Service
}

func NewProductHandler(svc Service) *ProductHandler {
	return &ProductHandler{svc: svc}
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
	filter := ProductFilter{}
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filter.CategoryID = &categoryID
	}

	products, err := h.svc.ListProducts(r.Context(), filter)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	product, err := h.svc.GetProductBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
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

	product, err := h.svc.CreateProduct(r.Context(), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.UpdateProduct(r.Context(), chi.URLParam(r, "slug"), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, product)
}

func (h *ProductHandler) DeleteBySlug(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteProduct(r.Context(), chi.URLParam(r, "slug")); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
