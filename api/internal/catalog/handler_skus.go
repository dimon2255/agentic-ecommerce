package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type SKUHandler struct {
	svc Service
}

func NewSKUHandler(svc Service) *SKUHandler {
	return &SKUHandler{svc: svc}
}

func (h *SKUHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{skuId}", h.Delete)
	return r
}

func (h *SKUHandler) List(w http.ResponseWriter, r *http.Request) {
	skus, err := h.svc.ListSKUsWithAttributes(r.Context(), chi.URLParam(r, "productId"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, skus)
}

func (h *SKUHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sku, err := h.svc.CreateSKUWithAttributes(r.Context(), chi.URLParam(r, "productId"), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, sku)
}

func (h *SKUHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteSKU(r.Context(), chi.URLParam(r, "skuId")); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
