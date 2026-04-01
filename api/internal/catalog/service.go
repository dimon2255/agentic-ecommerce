package catalog

import "context"

// Service defines business operations for the catalog domain.
type Service interface {
	// Categories
	ListCategories(ctx context.Context, filter CategoryFilter) ([]Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	CreateCategory(ctx context.Context, req CreateCategoryRequest) (*Category, error)
	UpdateCategory(ctx context.Context, slug string, req UpdateCategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, slug string) error

	// Products
	ListProducts(ctx context.Context, filter ProductFilter) ([]Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*Product, error)
	CreateProduct(ctx context.Context, req CreateProductRequest) (*Product, error)
	UpdateProduct(ctx context.Context, slug string, req UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, slug string) error

	// Attributes
	ListAttributesWithOptions(ctx context.Context, categoryID string) ([]CategoryAttribute, error)
	CreateAttribute(ctx context.Context, categoryID string, req CreateAttributeRequest) (*CategoryAttribute, error)
	DeleteAttribute(ctx context.Context, attrID string) error

	// Attribute Options
	ListOptions(ctx context.Context, attrID string) ([]AttributeOption, error)
	CreateOption(ctx context.Context, attrID string, req CreateAttributeOptionRequest) (*AttributeOption, error)
	DeleteOption(ctx context.Context, optionID string) error

	// SKUs
	ListSKUsWithAttributes(ctx context.Context, productID string) ([]SKU, error)
	CreateSKUWithAttributes(ctx context.Context, productID string, req CreateSKURequest) (*SKU, error)
	DeleteSKU(ctx context.Context, skuID string) error

	// Custom Fields
	ListCustomFields(ctx context.Context, entityType, entityID string) ([]CustomField, error)
	CreateCustomField(ctx context.Context, entityType, entityID string, req CreateCustomFieldRequest) (*CustomField, error)
	DeleteCustomField(ctx context.Context, fieldID string) error
}
