package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CustomFieldHandler struct {
	db *supabase.Client
}

func NewCustomFieldHandler(db *supabase.Client) *CustomFieldHandler {
	return &CustomFieldHandler{db: db}
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

	var fields []CustomField
	err := h.db.From("custom_fields").
		Select("*").
		Eq("entity_type", entityType).
		Eq("entity_id", entityID).
		Execute(&fields)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch custom fields")
		return
	}

	if fields == nil {
		fields = []CustomField{}
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

	if req.Key == "" || req.Value == "" {
		response.Error(w, http.StatusBadRequest, "key and value are required")
		return
	}

	insertData := map[string]any{
		"entity_type": entityType,
		"entity_id":   entityID,
		"key":         req.Key,
		"value":       req.Value,
	}

	var created []CustomField
	if err := h.db.From("custom_fields").Insert(insertData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create custom field")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *CustomFieldHandler) Delete(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	err := h.db.From("custom_fields").Eq("id", fieldID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete custom field")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
