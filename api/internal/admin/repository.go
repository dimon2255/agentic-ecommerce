package admin

import "context"

// Repository defines admin-specific data access operations.
// Catalog CRUD is delegated to the existing catalog.Service.
type Repository interface {
	// Orders — cross-user listing (unlike checkout repo which is user-scoped)
	ListOrders(ctx context.Context, filter OrderFilter) ([]OrderSummary, int, error)
	GetOrder(ctx context.Context, orderID string) (*OrderDetail, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error

	// Reports — read from PostgreSQL views
	GetDashboardKPIs(ctx context.Context) (*DashboardKPIs, error)
	GetSalesByDay(ctx context.Context, dateFrom, dateTo string) ([]SalesByDay, error)
	GetTokenUsageByDay(ctx context.Context, dateFrom, dateTo string) ([]TokenUsageByDay, error)

	// Audit log
	ListAuditLog(ctx context.Context, filter AuditLogFilter) ([]AuditLogEntry, int, error)
}
