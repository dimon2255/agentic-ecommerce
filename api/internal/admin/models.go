package admin

import (
	"strings"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
	"github.com/dimon2255/agentic-ecommerce/api/internal/validate"
)

var allowedImageContentTypes = []string{
	"image/jpeg", "image/png", "image/webp", "image/gif",
}

// --- Orders (admin view) ---

type OrderSummary struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id"`
	Status    string    `json:"status"`
	Email     string    `json:"email"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderDetail struct {
	ID                    string      `json:"id"`
	UserID                *string     `json:"user_id"`
	Status                string      `json:"status"`
	Email                 string      `json:"email"`
	ShippingAddress       any         `json:"shipping_address"`
	Subtotal              float64     `json:"subtotal"`
	Total                 float64     `json:"total"`
	StripePaymentIntentID *string     `json:"stripe_payment_intent_id"`
	Items                 []OrderItem `json:"items"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	SKUID       string  `json:"sku_id"`
	ProductName string  `json:"product_name"`
	SKUCode     string  `json:"sku_code"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type OrderFilter struct {
	Status   string
	Search   string // by order ID or email
	DateFrom string // YYYY-MM-DD
	DateTo   string // YYYY-MM-DD
	SortBy   string
	SortDir  string
	pagination.Params
}

var validOrderStatuses = []string{"draft", "pending", "paid", "shipped", "completed", "cancelled"}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (r *UpdateOrderStatusRequest) Validate() error {
	v := validate.New()
	v.Required("status", r.Status)
	v.OneOf("status", r.Status, validOrderStatuses)
	return v.Validate()
}

// --- Reports ---

type DashboardKPIs struct {
	TotalOrders    int     `json:"total_orders"`
	TotalRevenue   float64 `json:"total_revenue"`
	ActiveProducts int     `json:"active_products"`
	TotalCustomers int     `json:"total_customers"`
}

type SalesByDay struct {
	Day        string  `json:"day"`
	OrderCount int     `json:"order_count"`
	Revenue    float64 `json:"revenue"`
}

type TokenUsageByDay struct {
	Day          string `json:"day"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	RequestCount int    `json:"request_count"`
}

// --- Audit Log ---

type AuditLogEntry struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   *string   `json:"resource_id"`
	Changes      any       `json:"changes"`
	IPAddress    *string   `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuditLogFilter struct {
	Action       string
	ResourceType string
	UserID       string
	DateFrom     string
	DateTo       string
	pagination.Params
}

// --- Images ---

type UploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

func (r *UploadURLRequest) Validate() error {
	v := validate.New()
	v.Required("filename", r.Filename)
	v.MaxLength("filename", r.Filename, 255)
	v.Required("content_type", r.ContentType)
	v.OneOf("content_type", r.ContentType, allowedImageContentTypes)

	// Reject path traversal sequences
	if strings.Contains(r.Filename, "..") ||
		strings.Contains(r.Filename, "/") ||
		strings.Contains(r.Filename, "\\") ||
		strings.Contains(r.Filename, "\x00") {
		v.AddError("filename", "contains invalid characters")
	}

	return v.Validate()
}

type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
}
