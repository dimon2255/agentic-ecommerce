package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type CustomFieldHandler struct {
	svc Service
}

func NewCustomFieldHandler(svc Service) *CustomFieldHandler {
	return &CustomFieldHandler{svc: svc}
}

func (h *CustomFieldHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{fieldId}", h.Delete)
	return r
}

func (h *CustomFieldHandler) List(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	if entityType == "" || entityID == "" {
		response.Error(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	fields, err := h.svc.ListCustomFields(r.Context(), entityType, entityID)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, fields)
}

func (h *CustomFieldHandler) Create(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	if entityType == "" || entityID == "" {
		response.Error(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	var req CreateCustomFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	field, err := h.svc.CreateCustomField(r.Context(), entityType, entityID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, field)
}

func (h *CustomFieldHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteCustomField(r.Context(), chi.URLParam(r, "fieldId")); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
