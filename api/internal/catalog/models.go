package catalog

import "time"

// --- Categories ---

type Category struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	ParentID  *string    `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ParentID *string `json:"parent_id,omitempty"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name,omitempty"`
	Slug     *string `json:"slug,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}

// --- Category Attributes ---

type CategoryAttribute struct {
	ID         string            `json:"id"`
	CategoryID string            `json:"category_id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Required   bool              `json:"required"`
	SortOrder  int               `json:"sort_order"`
	Options    []AttributeOption `json:"options,omitempty"`
}

type CreateAttributeRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	SortOrder int    `json:"sort_order"`
}

type AttributeOption struct {
	ID                  string `json:"id"`
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
	SortOrder           int    `json:"sort_order"`
}

type CreateAttributeOptionRequest struct {
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// --- Products ---

type Product struct {
	ID          string    `json:"id"`
	CategoryID  string    `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	BasePrice   float64   `json:"base_price"`
	Status      string    `json:"status"`
	Images      []string  `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	CategoryID  string   `json:"category_id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description *string  `json:"description,omitempty"`
	BasePrice   float64  `json:"base_price"`
	Status      string   `json:"status"`
	Images      []string `json:"images,omitempty"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Slug        *string  `json:"slug,omitempty"`
	Description *string  `json:"description,omitempty"`
	BasePrice   *float64 `json:"base_price,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Images      []string `json:"images,omitempty"`
}

// --- SKUs ---

type SKU struct {
	ID              string              `json:"id"`
	ProductID       string              `json:"product_id"`
	SKUCode         string              `json:"sku_code"`
	PriceOverride   *float64            `json:"price_override"`
	Status          string              `json:"status"`
	AttributeValues []SKUAttributeValue `json:"attribute_values,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type CreateSKURequest struct {
	SKUCode         string                        `json:"sku_code"`
	PriceOverride   *float64                      `json:"price_override,omitempty"`
	Status          string                        `json:"status"`
	AttributeValues []CreateSKUAttributeValueReq  `json:"attribute_values"`
}

type CreateSKUAttributeValueReq struct {
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
}

type SKUAttributeValue struct {
	ID                  string `json:"id"`
	SKUID               string `json:"sku_id"`
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
}

// --- Custom Fields ---

type CustomField struct {
	ID         string `json:"id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Key        string `json:"key"`
	Value      string `json:"value"`
}

type CreateCustomFieldRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
