package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type SKUHandler struct {
	db *supabase.Client
}

func NewSKUHandler(db *supabase.Client) *SKUHandler {
	return &SKUHandler{db: db}
}

func (h *SKUHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{skuId}", h.Delete)
	return r
}

func (h *SKUHandler) List(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")

	var skus []SKU
	err := h.db.From("skus").
		Select("*").
		Eq("product_id", productID).
		Order("created_at", "asc").
		Execute(&skus)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch SKUs")
		return
	}

	// Fetch attribute values for each SKU
	for i := range skus {
		var attrValues []SKUAttributeValue
		h.db.From("sku_attribute_values").
			Select("*").
			Eq("sku_id", skus[i].ID).
			Execute(&attrValues)
		if attrValues == nil {
			attrValues = []SKUAttributeValue{}
		}
		skus[i].AttributeValues = attrValues
	}

	if skus == nil {
		skus = []SKU{}
	}
	response.JSON(w, http.StatusOK, skus)
}

func (h *SKUHandler) Create(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")

	var req CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SKUCode == "" {
		response.Error(w, http.StatusBadRequest, "sku_code is required")
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	// Insert SKU
	skuData := map[string]any{
		"product_id":     productID,
		"sku_code":       req.SKUCode,
		"price_override": req.PriceOverride,
		"status":         req.Status,
	}

	var created []SKU
	if err := h.db.From("skus").Insert(skuData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create SKU")
		return
	}

	sku := created[0]

	// Insert attribute values
	for _, av := range req.AttributeValues {
		avData := map[string]any{
			"sku_id":                sku.ID,
			"category_attribute_id": av.CategoryAttributeID,
			"value":                 av.Value,
		}
		h.db.From("sku_attribute_values").Insert(avData).Execute(nil)
	}

	// Fetch the attribute values back for the response
	var attrValues []SKUAttributeValue
	h.db.From("sku_attribute_values").Select("*").Eq("sku_id", sku.ID).Execute(&attrValues)
	sku.AttributeValues = attrValues

	response.JSON(w, http.StatusCreated, sku)
}

func (h *SKUHandler) Delete(w http.ResponseWriter, r *http.Request) {
	skuID := chi.URLParam(r, "skuId")

	err := h.db.From("skus").Eq("id", skuID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete SKU")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
