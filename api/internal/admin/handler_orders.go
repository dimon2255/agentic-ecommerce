package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// OrderHandler handles admin order management.
type OrderHandler struct {
	repo  Repository
	audit *AuditService
}

func NewOrderHandler(repo Repository, audit *AuditService) *OrderHandler {
	return &OrderHandler{repo: repo, audit: audit}
}

func (h *OrderHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}/status", h.UpdateStatus)
	return r
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := OrderFilter{Params: pagination.ParseFromQuery(r)}
	filter.Status = r.URL.Query().Get("status")
	filter.Search = r.URL.Query().Get("search")
	filter.DateFrom = r.URL.Query().Get("date_from")
	filter.DateTo = r.URL.Query().Get("date_to")
	filter.SortBy = r.URL.Query().Get("sort_by")
	filter.SortDir = r.URL.Query().Get("sort_dir")

	orders, total, err := h.repo.ListOrders(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}
	response.JSON(w, http.StatusOK, pagination.NewResponse(orders, total, filter.Params))
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	order, err := h.repo.GetOrder(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusNotFound, "order not found")
		return
	}
	response.JSON(w, http.StatusOK, order)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}

	// Get current order for audit trail
	order, err := h.repo.GetOrder(r.Context(), orderID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "order not found")
		return
	}

	oldStatus := order.Status
	if err := h.repo.UpdateOrderStatus(r.Context(), orderID, req.Status); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	h.audit.LogFromRequest(r, userID, "order:update_status", "order", orderID, map[string]string{
		"old_status": oldStatus,
		"new_status": req.Status,
	})

	response.JSON(w, http.StatusOK, map[string]string{
		"id":     orderID,
		"status": req.Status,
	})
}
