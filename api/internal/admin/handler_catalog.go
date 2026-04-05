package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// CatalogHandler wraps the existing catalog.Service for admin use, adding audit logging.
type CatalogHandler struct {
	svc   catalog.Service
	audit *AuditService
}

func NewCatalogHandler(svc catalog.Service, audit *AuditService) *CatalogHandler {
	return &CatalogHandler{svc: svc, audit: audit}
}

func (h *CatalogHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Products
	r.Get("/products", h.ListProducts)
	r.Post("/products", h.CreateProduct)
	r.Get("/products/{slug}", h.GetProduct)
	r.Patch("/products/{slug}", h.UpdateProduct)
	r.Delete("/products/{slug}", h.DeleteProduct)

	// SKUs under products
	r.Get("/products/{productId}/skus", h.ListSKUs)
	r.Post("/products/{productId}/skus", h.CreateSKU)
	r.Delete("/skus/{skuId}", h.DeleteSKU)

	// Categories
	r.Get("/categories", h.ListCategories)
	r.Post("/categories", h.CreateCategory)
	r.Get("/categories/{slug}", h.GetCategory)
	r.Patch("/categories/{slug}", h.UpdateCategory)
	r.Delete("/categories/{slug}", h.DeleteCategory)

	// Attributes under categories
	r.Get("/categories/{categoryId}/attributes", h.ListAttributes)
	r.Post("/categories/{categoryId}/attributes", h.CreateAttribute)
	r.Delete("/attributes/{attrId}", h.DeleteAttribute)

	// Attribute options
	r.Get("/attributes/{attrId}/options", h.ListOptions)
	r.Post("/attributes/{attrId}/options", h.CreateOption)
	r.Delete("/options/{optionId}", h.DeleteOption)

	return r
}

// --- Products ---

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	filter := catalog.ProductFilter{Params: pagination.ParseFromQuery(r)}
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filter.CategoryID = &categoryID
	}
	if categoryIDs := r.URL.Query().Get("category_ids"); categoryIDs != "" {
		filter.CategoryIDs = strings.Split(categoryIDs, ",")
	}
	filter.Search = r.URL.Query().Get("search")
	filter.SortBy = r.URL.Query().Get("sort_by")
	filter.SortDir = r.URL.Query().Get("sort_dir")

	products, total, err := h.svc.ListProducts(r.Context(), filter)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, pagination.NewResponse(products, total, filter.Params))
}

func (h *CatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.svc.GetProductBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, product)
}

func (h *CatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req catalog.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.CreateProduct(r.Context(), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "product:create", "product", product.ID, req)
	response.JSON(w, http.StatusCreated, product)
}

func (h *CatalogHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var req catalog.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.UpdateProduct(r.Context(), slug, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "product:update", "product", product.ID, req)
	response.JSON(w, http.StatusOK, product)
}

func (h *CatalogHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Get product ID for audit before deletion
	product, err := h.svc.GetProductBySlug(r.Context(), slug)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	if err := h.svc.DeleteProduct(r.Context(), slug); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "product:delete", "product", product.ID, map[string]string{"slug": slug})
	w.WriteHeader(http.StatusNoContent)
}

// --- SKUs ---

func (h *CatalogHandler) ListSKUs(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	skus, err := h.svc.ListSKUsWithAttributes(r.Context(), productID)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, skus)
}

func (h *CatalogHandler) CreateSKU(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	var req catalog.CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sku, err := h.svc.CreateSKUWithAttributes(r.Context(), productID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "sku:create", "sku", sku.ID, req)
	response.JSON(w, http.StatusCreated, sku)
}

func (h *CatalogHandler) DeleteSKU(w http.ResponseWriter, r *http.Request) {
	skuID := chi.URLParam(r, "skuId")
	if err := h.svc.DeleteSKU(r.Context(), skuID); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "sku:delete", "sku", skuID, nil)
	w.WriteHeader(http.StatusNoContent)
}

// --- Categories ---

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	filter := catalog.CategoryFilter{Params: pagination.ParseFromQuery(r)}
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

func (h *CatalogHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	cat, err := h.svc.GetCategoryBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, cat)
}

func (h *CatalogHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req catalog.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "category:create", "category", cat.ID, req)
	response.JSON(w, http.StatusCreated, cat)
}

func (h *CatalogHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var req catalog.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.UpdateCategory(r.Context(), slug, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "category:update", "category", cat.ID, req)
	response.JSON(w, http.StatusOK, cat)
}

func (h *CatalogHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	cat, err := h.svc.GetCategoryBySlug(r.Context(), slug)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	if err := h.svc.DeleteCategory(r.Context(), slug); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "category:delete", "category", cat.ID, map[string]string{"slug": slug})
	w.WriteHeader(http.StatusNoContent)
}

// --- Attributes ---

func (h *CatalogHandler) ListAttributes(w http.ResponseWriter, r *http.Request) {
	attrs, err := h.svc.ListAttributesWithOptions(r.Context(), chi.URLParam(r, "categoryId"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, attrs)
}

func (h *CatalogHandler) CreateAttribute(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")
	var req catalog.CreateAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	attr, err := h.svc.CreateAttribute(r.Context(), categoryID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "attribute:create", "attribute", attr.ID, req)
	response.JSON(w, http.StatusCreated, attr)
}

func (h *CatalogHandler) DeleteAttribute(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")
	if err := h.svc.DeleteAttribute(r.Context(), attrID); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "attribute:delete", "attribute", attrID, nil)
	w.WriteHeader(http.StatusNoContent)
}

// --- Attribute Options ---

func (h *CatalogHandler) ListOptions(w http.ResponseWriter, r *http.Request) {
	options, err := h.svc.ListOptions(r.Context(), chi.URLParam(r, "attrId"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, options)
}

func (h *CatalogHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")
	var req catalog.CreateAttributeOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	option, err := h.svc.CreateOption(r.Context(), attrID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "option:create", "attribute_option", option.ID, req)
	response.JSON(w, http.StatusCreated, option)
}

func (h *CatalogHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "optionId")
	if err := h.svc.DeleteOption(r.Context(), optionID); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "option:delete", "attribute_option", optionID, nil)
	w.WriteHeader(http.StatusNoContent)
}
