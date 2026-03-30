package cart

import "time"

// --- Database Models ---

type Cart struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id"`
	SessionID string    `json:"session_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartItem struct {
	ID        string    `json:"id"`
	CartID    string    `json:"cart_id"`
	SKUID     string    `json:"sku_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartItemWithSKU includes nested SKU/product data from PostgREST embedded select.
// PostgREST uses table names as JSON keys for embedded resources.
type CartItemWithSKU struct {
	ID        string    `json:"id"`
	CartID    string    `json:"cart_id"`
	SKUID     string    `json:"sku_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SKU       SKUEmbed  `json:"skus"`
}

type SKUEmbed struct {
	SKUCode       string       `json:"sku_code"`
	PriceOverride *float64     `json:"price_override"`
	Product       ProductEmbed `json:"products"`
}

type ProductEmbed struct {
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	BasePrice float64  `json:"base_price"`
	Images    []string `json:"images"`
}

// SKUForPrice is used when looking up current SKU price for cart snapshot.
type SKUForPrice struct {
	PriceOverride *float64     `json:"price_override"`
	Product       ProductEmbed `json:"products"`
}

// --- Request/Response Types ---

type CartResponse struct {
	ID    string            `json:"id"`
	Items []CartItemWithSKU `json:"items"`
}

type AddItemRequest struct {
	SKUID    string `json:"sku_id"`
	Quantity int    `json:"quantity"`
}

type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

type MergeCartRequest struct {
	SessionID string `json:"session_id"`
}
