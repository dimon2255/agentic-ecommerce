package checkout

import "time"

// Database models

type Order struct {
	ID                    string    `json:"id"`
	UserID                *string   `json:"user_id"`
	Status                string    `json:"status"`
	Email                 string    `json:"email"`
	ShippingAddress       any       `json:"shipping_address"`
	Subtotal              float64   `json:"subtotal"`
	Total                 float64   `json:"total"`
	StripePaymentIntentID *string   `json:"stripe_payment_intent_id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
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

// Request/Response types

type ShippingAddress struct {
	Name    string `json:"name"`
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state,omitempty"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

type StartCheckoutRequest struct {
	Email           string          `json:"email"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
}

type StartCheckoutResponse struct {
	OrderID      string `json:"order_id"`
	ClientSecret string `json:"client_secret"`
}

type PriceChange struct {
	SKUID    string  `json:"sku_id"`
	SKUCode  string  `json:"sku_code"`
	OldPrice float64 `json:"old_price"`
	NewPrice float64 `json:"new_price"`
}

type OrderResponse struct {
	ID              string              `json:"id"`
	Status          string              `json:"status"`
	Email           string              `json:"email"`
	ShippingAddress any                 `json:"shipping_address"`
	Subtotal        float64             `json:"subtotal"`
	Total           float64             `json:"total"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
}

type OrderItemResponse struct {
	ProductName string  `json:"product_name"`
	SKUCode     string  `json:"sku_code"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// Internal types for cart data deserialization (PostgREST embedded select)

type cartItem struct {
	ID        string   `json:"id"`
	SKUID     string   `json:"sku_id"`
	Quantity  int      `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	SKU       skuEmbed `json:"skus"`
}

type skuEmbed struct {
	SKUCode       string       `json:"sku_code"`
	PriceOverride *float64     `json:"price_override"`
	Product       productEmbed `json:"products"`
}

type productEmbed struct {
	Name      string  `json:"name"`
	BasePrice float64 `json:"base_price"`
}

type cart struct {
	ID     string  `json:"id"`
	UserID *string `json:"user_id"`
	Status string  `json:"status"`
}
