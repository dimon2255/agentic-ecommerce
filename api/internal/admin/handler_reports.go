package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// ReportsHandler serves admin dashboard and report data.
type ReportsHandler struct {
	repo Repository
}

func NewReportsHandler(repo Repository) *ReportsHandler {
	return &ReportsHandler{repo: repo}
}

func (h *ReportsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/dashboard", h.Dashboard)
	r.Get("/sales", h.SalesByDay)
	r.Get("/token-usage", h.TokenUsageByDay)
	return r
}

func (h *ReportsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	kpis, err := h.repo.GetDashboardKPIs(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch dashboard data")
		return
	}
	response.JSON(w, http.StatusOK, kpis)
}

func (h *ReportsHandler) SalesByDay(w http.ResponseWriter, r *http.Request) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	sales, err := h.repo.GetSalesByDay(r.Context(), dateFrom, dateTo)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch sales data")
		return
	}
	response.JSON(w, http.StatusOK, sales)
}

func (h *ReportsHandler) TokenUsageByDay(w http.ResponseWriter, r *http.Request) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	usage, err := h.repo.GetTokenUsageByDay(r.Context(), dateFrom, dateTo)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch token usage data")
		return
	}
	response.JSON(w, http.StatusOK, usage)
}
