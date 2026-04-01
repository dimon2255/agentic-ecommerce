package catalog

import (
	"context"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
)

type catalogService struct {
	repo Repository
}

// NewService creates a catalog service backed by the given repository.
func NewService(repo Repository) Service {
	return &catalogService{repo: repo}
}

// --- Categories ---

func (s *catalogService) ListCategories(ctx context.Context, filter CategoryFilter) ([]Category, error) {
	cats, err := s.repo.ListCategories(ctx, filter)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch categories", err)
	}
	return cats, nil
}

func (s *catalogService) GetCategoryBySlug(ctx context.Context, slug string) (*Category, error) {
	cat, err := s.repo.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.NewNotFound("category")
	}
	return cat, nil
}

func (s *catalogService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*Category, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	cat, err := s.repo.CreateCategory(ctx, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to create category", err)
	}
	return cat, nil
}

func (s *catalogService) UpdateCategory(ctx context.Context, slug string, req UpdateCategoryRequest) (*Category, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	cat, err := s.repo.UpdateCategory(ctx, slug, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to update category", err)
	}
	if cat == nil {
		return nil, apperror.NewNotFound("category")
	}
	return cat, nil
}

func (s *catalogService) DeleteCategory(ctx context.Context, slug string) error {
	if err := s.repo.DeleteCategory(ctx, slug); err != nil {
		return apperror.NewInternal("failed to delete category", err)
	}
	return nil
}

// --- Products ---

func (s *catalogService) ListProducts(ctx context.Context, filter ProductFilter) ([]Product, error) {
	products, err := s.repo.ListProducts(ctx, filter)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch products", err)
	}
	return products, nil
}

func (s *catalogService) GetProductBySlug(ctx context.Context, slug string) (*Product, error) {
	product, err := s.repo.GetProductBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.NewNotFound("product")
	}
	return product, nil
}

func (s *catalogService) CreateProduct(ctx context.Context, req CreateProductRequest) (*Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	product, err := s.repo.CreateProduct(ctx, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to create product", err)
	}
	return product, nil
}

func (s *catalogService) UpdateProduct(ctx context.Context, slug string, req UpdateProductRequest) (*Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	product, err := s.repo.UpdateProduct(ctx, slug, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to update product", err)
	}
	if product == nil {
		return nil, apperror.NewNotFound("product")
	}
	return product, nil
}

func (s *catalogService) DeleteProduct(ctx context.Context, slug string) error {
	if err := s.repo.DeleteProduct(ctx, slug); err != nil {
		return apperror.NewInternal("failed to delete product", err)
	}
	return nil
}

// --- Attributes ---

func (s *catalogService) ListAttributesWithOptions(ctx context.Context, categoryID string) ([]CategoryAttribute, error) {
	attrs, err := s.repo.ListAttributesWithOptions(ctx, categoryID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch attributes", err)
	}
	return attrs, nil
}

func (s *catalogService) CreateAttribute(ctx context.Context, categoryID string, req CreateAttributeRequest) (*CategoryAttribute, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	attr, err := s.repo.CreateAttribute(ctx, categoryID, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to create attribute", err)
	}
	return attr, nil
}

func (s *catalogService) DeleteAttribute(ctx context.Context, attrID string) error {
	if err := s.repo.DeleteAttribute(ctx, attrID); err != nil {
		return apperror.NewInternal("failed to delete attribute", err)
	}
	return nil
}

// --- Attribute Options ---

func (s *catalogService) ListOptions(ctx context.Context, attrID string) ([]AttributeOption, error) {
	options, err := s.repo.ListOptions(ctx, attrID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch options", err)
	}
	return options, nil
}

func (s *catalogService) CreateOption(ctx context.Context, attrID string, req CreateAttributeOptionRequest) (*AttributeOption, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	option, err := s.repo.CreateOption(ctx, attrID, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to create option", err)
	}
	return option, nil
}

func (s *catalogService) DeleteOption(ctx context.Context, optionID string) error {
	if err := s.repo.DeleteOption(ctx, optionID); err != nil {
		return apperror.NewInternal("failed to delete option", err)
	}
	return nil
}

// --- SKUs ---

func (s *catalogService) ListSKUsWithAttributes(ctx context.Context, productID string) ([]SKU, error) {
	skus, err := s.repo.ListSKUsWithAttributes(ctx, productID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch SKUs", err)
	}
	return skus, nil
}

func (s *catalogService) CreateSKUWithAttributes(ctx context.Context, productID string, req CreateSKURequest) (*SKU, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	sku, err := s.repo.CreateSKUWithAttributes(ctx, productID, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to create SKU", err)
	}
	return sku, nil
}

func (s *catalogService) DeleteSKU(ctx context.Context, skuID string) error {
	if err := s.repo.DeleteSKU(ctx, skuID); err != nil {
		return apperror.NewInternal("failed to delete SKU", err)
	}
	return nil
}

// --- Custom Fields ---

func (s *catalogService) ListCustomFields(ctx context.Context, entityType, entityID string) ([]CustomField, error) {
	fields, err := s.repo.ListCustomFields(ctx, entityType, entityID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch custom fields", err)
	}
	return fields, nil
}

func (s *catalogService) CreateCustomField(ctx context.Context, entityType, entityID string, req CreateCustomFieldRequest) (*CustomField, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	field, err := s.repo.CreateCustomField(ctx, entityType, entityID, req)
	if err != nil {
		return nil, apperror.NewInternal("failed to create custom field", err)
	}
	return field, nil
}

func (s *catalogService) DeleteCustomField(ctx context.Context, fieldID string) error {
	if err := s.repo.DeleteCustomField(ctx, fieldID); err != nil {
		return apperror.NewInternal("failed to delete custom field", err)
	}
	return nil
}
