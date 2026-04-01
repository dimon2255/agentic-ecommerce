package catalog

import (
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/internal/validate"
)

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

func (r *CreateCategoryRequest) Validate() error {
	v := validate.New()
	v.Required("name", r.Name)
	v.Required("slug", r.Slug)
	v.MinLength("slug", r.Slug, 2)
	v.MaxLength("name", r.Name, 255)
	if r.ParentID != nil {
		v.UUID("parent_id", *r.ParentID)
	}
	return v.Validate()
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name,omitempty"`
	Slug     *string `json:"slug,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}

func (r *UpdateCategoryRequest) Validate() error {
	v := validate.New()
	if r.Name != nil {
		v.MinLength("name", *r.Name, 1)
		v.MaxLength("name", *r.Name, 255)
	}
	if r.Slug != nil {
		v.MinLength("slug", *r.Slug, 2)
	}
	if r.ParentID != nil {
		v.UUID("parent_id", *r.ParentID)
	}
	return v.Validate()
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

var attributeTypes = []string{"text", "number", "enum"}

func (r *CreateAttributeRequest) Validate() error {
	v := validate.New()
	v.Required("name", r.Name)
	v.Required("type", r.Type)
	v.OneOf("type", r.Type, attributeTypes)
	return v.Validate()
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

func (r *CreateAttributeOptionRequest) Validate() error {
	v := validate.New()
	v.Required("value", r.Value)
	return v.Validate()
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

var productStatuses = []string{"draft", "active", "archived"}

func (r *CreateProductRequest) Validate() error {
	v := validate.New()
	v.Required("category_id", r.CategoryID)
	v.UUID("category_id", r.CategoryID)
	v.Required("name", r.Name)
	v.Required("slug", r.Slug)
	v.MinLength("slug", r.Slug, 2)
	v.MaxLength("name", r.Name, 255)
	v.FloatMin("base_price", r.BasePrice, 0)
	if r.Status == "" {
		r.Status = "draft"
	}
	v.OneOf("status", r.Status, productStatuses)
	return v.Validate()
}

type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Slug        *string  `json:"slug,omitempty"`
	Description *string  `json:"description,omitempty"`
	BasePrice   *float64 `json:"base_price,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Images      []string `json:"images,omitempty"`
}

func (r *UpdateProductRequest) Validate() error {
	v := validate.New()
	if r.Name != nil {
		v.MinLength("name", *r.Name, 1)
		v.MaxLength("name", *r.Name, 255)
	}
	if r.Slug != nil {
		v.MinLength("slug", *r.Slug, 2)
	}
	if r.BasePrice != nil {
		v.FloatMin("base_price", *r.BasePrice, 0)
	}
	if r.Status != nil {
		v.OneOf("status", *r.Status, productStatuses)
	}
	return v.Validate()
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

var skuStatuses = []string{"active", "inactive"}

func (r *CreateSKURequest) Validate() error {
	v := validate.New()
	v.Required("sku_code", r.SKUCode)
	v.MinLength("sku_code", r.SKUCode, 2)
	if r.PriceOverride != nil {
		v.FloatMin("price_override", *r.PriceOverride, 0)
	}
	if r.Status == "" {
		r.Status = "active"
	}
	v.OneOf("status", r.Status, skuStatuses)
	return v.Validate()
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

func (r *CreateCustomFieldRequest) Validate() error {
	v := validate.New()
	v.Required("key", r.Key)
	v.Required("value", r.Value)
	v.MaxLength("key", r.Key, 255)
	return v.Validate()
}
