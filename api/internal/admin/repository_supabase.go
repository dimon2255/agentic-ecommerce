package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

// --- Input validation helpers ---

var allowedOrderSortColumns = map[string]bool{
	"created_at": true, "email": true, "status": true, "total": true,
}

func validSort(col string, allowed map[string]bool, fallback string) string {
	if allowed[col] {
		return col
	}
	return fallback
}

func validSortDir(dir string) string {
	if dir == "asc" {
		return "asc"
	}
	return "desc"
}

func validDateParam(s string) string {
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return s
	}
	return ""
}

type supabaseRepository struct {
	db *supabase.Client
}

// NewSupabaseRepository creates an admin repository backed by Supabase PostgREST.
func NewSupabaseRepository(db *supabase.Client) Repository {
	return &supabaseRepository{db: db}
}

// --- Orders ---

func (r *supabaseRepository) ListOrders(_ context.Context, filter OrderFilter) ([]OrderSummary, int, error) {
	query := r.db.From("orders").
		Select("id,user_id,status,email,total,created_at").
		CountExact()

	// Exclude soft-deleted
	query = query.Is("deleted_at", "null")

	if filter.Status != "" {
		query = query.Eq("status", filter.Status)
	}
	if filter.Search != "" {
		query = query.Ilike("email", filter.Search)
	}
	if df := validDateParam(filter.DateFrom); df != "" {
		query = query.Gte("created_at", df)
	}
	if dt := validDateParam(filter.DateTo); dt != "" {
		query = query.Lte("created_at", dt+"T23:59:59Z")
	}

	query = query.Order(
		validSort(filter.SortBy, allowedOrderSortColumns, "created_at"),
		validSortDir(filter.SortDir),
	)

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(filter.Offset())
	}

	var orders []OrderSummary
	total, err := query.ExecuteWithCount(&orders)
	if err != nil {
		return nil, 0, err
	}
	if orders == nil {
		orders = []OrderSummary{}
	}
	return orders, total, nil
}

func (r *supabaseRepository) GetOrder(_ context.Context, orderID string) (*OrderDetail, error) {
	// Fetch order
	var orders []OrderDetail
	err := r.db.From("orders").
		Select("*").
		Eq("id", orderID).
		Is("deleted_at", "null").
		Limit(1).
		Execute(&orders)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("order not found")
	}
	order := &orders[0]

	// Fetch order items
	var items []OrderItem
	err = r.db.From("order_items").
		Select("*").
		Eq("order_id", orderID).
		Execute(&items)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []OrderItem{}
	}
	order.Items = items

	return order, nil
}

func (r *supabaseRepository) UpdateOrderStatus(_ context.Context, orderID, status string) error {
	return r.db.From("orders").
		Eq("id", orderID).
		Update(map[string]string{"status": status}).
		Execute(nil)
}

// --- Reports ---

func (r *supabaseRepository) GetDashboardKPIs(_ context.Context) (*DashboardKPIs, error) {
	var kpis []DashboardKPIs
	err := r.db.From("admin_dashboard_kpis").
		Select("*").
		Execute(&kpis)
	if err != nil {
		return nil, err
	}
	if len(kpis) == 0 {
		return &DashboardKPIs{}, nil
	}
	return &kpis[0], nil
}

func (r *supabaseRepository) GetSalesByDay(_ context.Context, dateFrom, dateTo string) ([]SalesByDay, error) {
	query := r.db.From("admin_sales_by_day").Select("*")
	if df := validDateParam(dateFrom); df != "" {
		query = query.Gte("day", df)
	}
	if dt := validDateParam(dateTo); dt != "" {
		query = query.Lte("day", dt)
	}
	query = query.Order("day", "desc").Limit(90)

	var sales []SalesByDay
	err := query.Execute(&sales)
	if err != nil {
		return nil, err
	}
	if sales == nil {
		sales = []SalesByDay{}
	}
	return sales, nil
}

func (r *supabaseRepository) GetTokenUsageByDay(_ context.Context, dateFrom, dateTo string) ([]TokenUsageByDay, error) {
	query := r.db.From("admin_token_usage_by_day").Select("*")
	if df := validDateParam(dateFrom); df != "" {
		query = query.Gte("day", df)
	}
	if dt := validDateParam(dateTo); dt != "" {
		query = query.Lte("day", dt)
	}
	query = query.Order("day", "desc").Limit(90)

	var usage []TokenUsageByDay
	err := query.Execute(&usage)
	if err != nil {
		return nil, err
	}
	if usage == nil {
		usage = []TokenUsageByDay{}
	}
	return usage, nil
}

// --- Audit Log ---

func (r *supabaseRepository) ListAuditLog(_ context.Context, filter AuditLogFilter) ([]AuditLogEntry, int, error) {
	query := r.db.From("admin_audit_log").
		Select("*").
		CountExact().
		Order("created_at", "desc")

	if filter.Action != "" {
		query = query.Eq("action", filter.Action)
	}
	if filter.ResourceType != "" {
		query = query.Eq("resource_type", filter.ResourceType)
	}
	if filter.UserID != "" {
		query = query.Eq("user_id", filter.UserID)
	}
	if df := validDateParam(filter.DateFrom); df != "" {
		query = query.Gte("created_at", df)
	}
	if dt := validDateParam(filter.DateTo); dt != "" {
		query = query.Lte("created_at", dt+"T23:59:59Z")
	}

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(filter.Offset())
	}

	var entries []AuditLogEntry
	total, err := query.ExecuteWithCount(&entries)
	if err != nil {
		return nil, 0, err
	}
	if entries == nil {
		entries = []AuditLogEntry{}
	}
	return entries, total, nil
}
