package catalog

import (
	"context"

	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
)

// Repository defines data access operations for the catalog domain.
type Repository interface {
	// Categories
	ListCategories(ctx context.Context, filter CategoryFilter) ([]Category, int, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	CreateCategory(ctx context.Context, data CreateCategoryRequest) (*Category, error)
	UpdateCategory(ctx context.Context, slug string, data UpdateCategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, slug string) error

	// Products
	ListProducts(ctx context.Context, filter ProductFilter) ([]Product, int, error)
	GetProductBySlug(ctx context.Context, slug string) (*Product, error)
	CreateProduct(ctx context.Context, data CreateProductRequest) (*Product, error)
	UpdateProduct(ctx context.Context, slug string, data UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, slug string) error

	// Attributes (returns attributes with embedded options — no N+1)
	ListAttributesWithOptions(ctx context.Context, categoryID string) ([]CategoryAttribute, error)
	CreateAttribute(ctx context.Context, categoryID string, data CreateAttributeRequest) (*CategoryAttribute, error)
	DeleteAttribute(ctx context.Context, attrID string) error

	// Attribute Options
	ListOptions(ctx context.Context, attrID string) ([]AttributeOption, error)
	CreateOption(ctx context.Context, attrID string, data CreateAttributeOptionRequest) (*AttributeOption, error)
	DeleteOption(ctx context.Context, optionID string) error

	// SKUs (returns SKUs with embedded attribute values — no N+1)
	ListSKUsWithAttributes(ctx context.Context, productID string) ([]SKU, error)
	CreateSKUWithAttributes(ctx context.Context, productID string, data CreateSKURequest) (*SKU, error)
	DeleteSKU(ctx context.Context, skuID string) error

	// Custom Fields
	ListCustomFields(ctx context.Context, entityType, entityID string) ([]CustomField, error)
	CreateCustomField(ctx context.Context, entityType, entityID string, data CreateCustomFieldRequest) (*CustomField, error)
	DeleteCustomField(ctx context.Context, fieldID string) error
}

// CategoryFilter controls category list queries.
type CategoryFilter struct {
	ParentID *string // filter by parent_id; "null" for root categories
	pagination.Params
}

// ProductFilter controls product list queries.
type ProductFilter struct {
	CategoryID  *string
	Search      string
	SortBy      string
	SortDir     string
	CategoryIDs []string // for multi-category queries (frontend N+1 fix)
	pagination.Params
}
