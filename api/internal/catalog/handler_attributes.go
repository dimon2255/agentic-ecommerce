package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type AttributeHandler struct {
	svc Service
}

func NewAttributeHandler(svc Service) *AttributeHandler {
	return &AttributeHandler{svc: svc}
}

func (h *AttributeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{attrId}", h.Delete)
	r.Get("/{attrId}/options", h.ListOptions)
	r.Post("/{attrId}/options", h.CreateOption)
	r.Delete("/{attrId}/options/{optionId}", h.DeleteOption)
	return r
}

func (h *AttributeHandler) List(w http.ResponseWriter, r *http.Request) {
	attrs, err := h.svc.ListAttributesWithOptions(r.Context(), chi.URLParam(r, "categoryId"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, attrs)
}

func (h *AttributeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	attr, err := h.svc.CreateAttribute(r.Context(), chi.URLParam(r, "categoryId"), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, attr)
}

func (h *AttributeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteAttribute(r.Context(), chi.URLParam(r, "attrId")); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AttributeHandler) ListOptions(w http.ResponseWriter, r *http.Request) {
	options, err := h.svc.ListOptions(r.Context(), chi.URLParam(r, "attrId"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, options)
}

func (h *AttributeHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	var req CreateAttributeOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	option, err := h.svc.CreateOption(r.Context(), chi.URLParam(r, "attrId"), req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, option)
}

func (h *AttributeHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteOption(r.Context(), chi.URLParam(r, "optionId")); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
