package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type AttributeHandler struct {
	db *supabase.Client
}

func NewAttributeHandler(db *supabase.Client) *AttributeHandler {
	return &AttributeHandler{db: db}
}

func (h *AttributeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{attrId}", h.Delete)

	// Attribute options
	r.Get("/{attrId}/options", h.ListOptions)
	r.Post("/{attrId}/options", h.CreateOption)
	r.Delete("/{attrId}/options/{optionId}", h.DeleteOption)
	return r
}

func (h *AttributeHandler) List(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")

	var attrs []CategoryAttribute
	err := h.db.From("category_attributes").
		Select("*").
		Eq("category_id", categoryID).
		Order("sort_order", "asc").
		Execute(&attrs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch attributes")
		return
	}

	// Fetch options for each attribute
	for i := range attrs {
		var options []AttributeOption
		h.db.From("attribute_options").
			Select("*").
			Eq("category_attribute_id", attrs[i].ID).
			Order("sort_order", "asc").
			Execute(&options)
		if options == nil {
			options = []AttributeOption{}
		}
		attrs[i].Options = options
	}

	if attrs == nil {
		attrs = []CategoryAttribute{}
	}
	response.JSON(w, http.StatusOK, attrs)
}

func (h *AttributeHandler) Create(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")

	var req CreateAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Type == "" {
		response.Error(w, http.StatusBadRequest, "name and type are required")
		return
	}

	insertData := map[string]any{
		"category_id": categoryID,
		"name":        req.Name,
		"type":        req.Type,
		"required":    req.Required,
		"sort_order":  req.SortOrder,
	}

	var created []CategoryAttribute
	if err := h.db.From("category_attributes").Insert(insertData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create attribute")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *AttributeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")

	err := h.db.From("category_attributes").Eq("id", attrID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete attribute")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AttributeHandler) ListOptions(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")

	var options []AttributeOption
	err := h.db.From("attribute_options").
		Select("*").
		Eq("category_attribute_id", attrID).
		Order("sort_order", "asc").
		Execute(&options)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch options")
		return
	}

	if options == nil {
		options = []AttributeOption{}
	}
	response.JSON(w, http.StatusOK, options)
}

func (h *AttributeHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")

	var req CreateAttributeOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	insertData := map[string]any{
		"category_attribute_id": attrID,
		"value":                 req.Value,
		"sort_order":            req.SortOrder,
	}

	var created []AttributeOption
	if err := h.db.From("attribute_options").Insert(insertData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create option")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *AttributeHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "optionId")

	err := h.db.From("attribute_options").Eq("id", optionID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete option")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
