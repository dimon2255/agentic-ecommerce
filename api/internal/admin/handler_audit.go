package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// AuditLogHandler serves the admin audit log.
type AuditLogHandler struct {
	repo Repository
}

func NewAuditLogHandler(repo Repository) *AuditLogHandler {
	return &AuditLogHandler{repo: repo}
}

func (h *AuditLogHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	return r
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := AuditLogFilter{Params: pagination.ParseFromQuery(r)}
	filter.Action = r.URL.Query().Get("action")
	filter.ResourceType = r.URL.Query().Get("resource_type")
	filter.UserID = r.URL.Query().Get("user_id")
	filter.DateFrom = r.URL.Query().Get("date_from")
	filter.DateTo = r.URL.Query().Get("date_to")

	entries, total, err := h.repo.ListAuditLog(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch audit log")
		return
	}
	response.JSON(w, http.StatusOK, pagination.NewResponse(entries, total, filter.Params))
}
