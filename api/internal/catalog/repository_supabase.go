package catalog

import (
	"context"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type supabaseRepository struct {
	db *supabase.Client
}

// NewSupabaseRepository creates a catalog repository backed by Supabase PostgREST.
func NewSupabaseRepository(db *supabase.Client) Repository {
	return &supabaseRepository{db: db}
}

// --- Categories ---

func (r *supabaseRepository) ListCategories(_ context.Context, filter CategoryFilter) ([]Category, int, error) {
	query := r.db.From("categories").Select("*").Order("name", "asc").CountExact()
	if filter.ParentID != nil {
		if *filter.ParentID == "null" {
			query = query.Is("parent_id", "null")
		} else {
			query = query.Eq("parent_id", *filter.ParentID)
		}
	}
	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(filter.Offset())
	}
	var categories []Category
	total, err := query.ExecuteWithCount(&categories)
	if err != nil {
		return nil, 0, err
	}
	if categories == nil {
		categories = []Category{}
	}
	return categories, total, nil
}

func (r *supabaseRepository) GetCategoryBySlug(_ context.Context, slug string) (*Category, error) {
	var cat Category
	err := r.db.From("categories").Select("*").Eq("slug", slug).Single().Execute(&cat)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *supabaseRepository) CreateCategory(_ context.Context, data CreateCategoryRequest) (*Category, error) {
	var created []Category
	if err := r.db.From("categories").Insert(data).Execute(&created); err != nil {
		return nil, err
	}
	return &created[0], nil
}

func (r *supabaseRepository) UpdateCategory(_ context.Context, slug string, data UpdateCategoryRequest) (*Category, error) {
	var updated []Category
	if err := r.db.From("categories").Eq("slug", slug).Update(data).Execute(&updated); err != nil {
		return nil, err
	}
	if len(updated) == 0 {
		return nil, nil // not found
	}
	return &updated[0], nil
}

func (r *supabaseRepository) DeleteCategory(_ context.Context, slug string) error {
	return r.db.From("categories").Eq("slug", slug).Delete().Execute(nil)
}

// --- Products ---

func (r *supabaseRepository) ListProducts(_ context.Context, filter ProductFilter) ([]Product, int, error) {
	// Determine sort
	sortBy := "created_at"
	sortDir := "desc"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortDir == "asc" || filter.SortDir == "desc" {
		sortDir = filter.SortDir
	}

	query := r.db.From("products").Select("*").Order(sortBy, sortDir).CountExact()

	// Filter by single category
	if filter.CategoryID != nil {
		query = query.Eq("category_id", *filter.CategoryID)
	}

	// Filter by multiple categories (for frontend N+1 fix)
	if len(filter.CategoryIDs) > 0 {
		query = query.In("category_id", filter.CategoryIDs)
	}

	// Full-text search
	if filter.Search != "" {
		query = query.Fts("search_vector", filter.Search)
	}

	// Pagination
	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage).Offset(filter.Offset())
	}

	var products []Product
	total, err := query.ExecuteWithCount(&products)
	if err != nil {
		return nil, 0, err
	}
	if products == nil {
		products = []Product{}
	}
	return products, total, nil
}

func (r *supabaseRepository) GetProductBySlug(_ context.Context, slug string) (*Product, error) {
	var product Product
	err := r.db.From("products").Select("*").Eq("slug", slug).Single().Execute(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *supabaseRepository) CreateProduct(_ context.Context, data CreateProductRequest) (*Product, error) {
	var created []Product
	if err := r.db.From("products").Insert(data).Execute(&created); err != nil {
		return nil, err
	}
	return &created[0], nil
}

func (r *supabaseRepository) UpdateProduct(_ context.Context, slug string, data UpdateProductRequest) (*Product, error) {
	var updated []Product
	if err := r.db.From("products").Eq("slug", slug).Update(data).Execute(&updated); err != nil {
		return nil, err
	}
	if len(updated) == 0 {
		return nil, nil
	}
	return &updated[0], nil
}

func (r *supabaseRepository) DeleteProduct(_ context.Context, slug string) error {
	return r.db.From("products").Eq("slug", slug).Delete().Execute(nil)
}

// --- Attributes (N+1 fix: embedded select) ---

// categoryAttributeRow maps PostgREST embedded select where table name is the JSON key.
type categoryAttributeRow struct {
	ID         string            `json:"id"`
	CategoryID string            `json:"category_id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Required   bool              `json:"required"`
	SortOrder  int               `json:"sort_order"`
	RawOptions []AttributeOption `json:"attribute_options"`
}

func (r *supabaseRepository) ListAttributesWithOptions(_ context.Context, categoryID string) ([]CategoryAttribute, error) {
	var rows []categoryAttributeRow
	err := r.db.From("category_attributes").
		Select("*,attribute_options(*)").
		Eq("category_id", categoryID).
		Order("sort_order", "asc").
		Execute(&rows)
	if err != nil {
		return nil, err
	}

	attrs := make([]CategoryAttribute, len(rows))
	for i, row := range rows {
		options := row.RawOptions
		if options == nil {
			options = []AttributeOption{}
		}
		attrs[i] = CategoryAttribute{
			ID:         row.ID,
			CategoryID: row.CategoryID,
			Name:       row.Name,
			Type:       row.Type,
			Required:   row.Required,
			SortOrder:  row.SortOrder,
			Options:    options,
		}
	}
	return attrs, nil
}

func (r *supabaseRepository) CreateAttribute(_ context.Context, categoryID string, data CreateAttributeRequest) (*CategoryAttribute, error) {
	insertData := map[string]any{
		"category_id": categoryID,
		"name":        data.Name,
		"type":        data.Type,
		"required":    data.Required,
		"sort_order":  data.SortOrder,
	}
	var created []CategoryAttribute
	if err := r.db.From("category_attributes").Insert(insertData).Execute(&created); err != nil {
		return nil, err
	}
	return &created[0], nil
}

func (r *supabaseRepository) DeleteAttribute(_ context.Context, attrID string) error {
	return r.db.From("category_attributes").Eq("id", attrID).Delete().Execute(nil)
}

// --- Attribute Options ---

func (r *supabaseRepository) ListOptions(_ context.Context, attrID string) ([]AttributeOption, error) {
	var options []AttributeOption
	err := r.db.From("attribute_options").
		Select("*").
		Eq("category_attribute_id", attrID).
		Order("sort_order", "asc").
		Execute(&options)
	if err != nil {
		return nil, err
	}
	if options == nil {
		options = []AttributeOption{}
	}
	return options, nil
}

func (r *supabaseRepository) CreateOption(_ context.Context, attrID string, data CreateAttributeOptionRequest) (*AttributeOption, error) {
	insertData := map[string]any{
		"category_attribute_id": attrID,
		"value":                 data.Value,
		"sort_order":            data.SortOrder,
	}
	var created []AttributeOption
	if err := r.db.From("attribute_options").Insert(insertData).Execute(&created); err != nil {
		return nil, err
	}
	return &created[0], nil
}

func (r *supabaseRepository) DeleteOption(_ context.Context, optionID string) error {
	return r.db.From("attribute_options").Eq("id", optionID).Delete().Execute(nil)
}

// --- SKUs (N+1 fix: embedded select) ---

// skuRow maps PostgREST embedded select for sku_attribute_values.
type skuRow struct {
	ID            string              `json:"id"`
	ProductID     string              `json:"product_id"`
	SKUCode       string              `json:"sku_code"`
	PriceOverride *float64            `json:"price_override"`
	Status        string              `json:"status"`
	RawAttrValues []SKUAttributeValue `json:"sku_attribute_values"`
	CreatedAt     interface{}         `json:"created_at"`
	UpdatedAt     interface{}         `json:"updated_at"`
}

func (r *supabaseRepository) ListSKUsWithAttributes(_ context.Context, productID string) ([]SKU, error) {
	var rows []skuRow
	err := r.db.From("skus").
		Select("*,sku_attribute_values(*)").
		Eq("product_id", productID).
		Order("created_at", "asc").
		Execute(&rows)
	if err != nil {
		return nil, err
	}

	skus := make([]SKU, len(rows))
	for i, row := range rows {
		attrValues := row.RawAttrValues
		if attrValues == nil {
			attrValues = []SKUAttributeValue{}
		}
		skus[i] = SKU{
			ID:              row.ID,
			ProductID:       row.ProductID,
			SKUCode:         row.SKUCode,
			PriceOverride:   row.PriceOverride,
			Status:          row.Status,
			AttributeValues: attrValues,
		}
	}
	return skus, nil
}

func (r *supabaseRepository) CreateSKUWithAttributes(_ context.Context, productID string, data CreateSKURequest) (*SKU, error) {
	skuData := map[string]any{
		"product_id":     productID,
		"sku_code":       data.SKUCode,
		"price_override": data.PriceOverride,
		"status":         data.Status,
	}
	var created []SKU
	if err := r.db.From("skus").Insert(skuData).Execute(&created); err != nil {
		return nil, err
	}
	sku := created[0]

	for _, av := range data.AttributeValues {
		avData := map[string]any{
			"sku_id":                sku.ID,
			"category_attribute_id": av.CategoryAttributeID,
			"value":                 av.Value,
		}
		r.db.From("sku_attribute_values").Insert(avData).Execute(nil)
	}

	var attrValues []SKUAttributeValue
	r.db.From("sku_attribute_values").Select("*").Eq("sku_id", sku.ID).Execute(&attrValues)
	sku.AttributeValues = attrValues

	return &sku, nil
}

func (r *supabaseRepository) DeleteSKU(_ context.Context, skuID string) error {
	return r.db.From("skus").Eq("id", skuID).Delete().Execute(nil)
}

// --- Custom Fields ---

func (r *supabaseRepository) ListCustomFields(_ context.Context, entityType, entityID string) ([]CustomField, error) {
	var fields []CustomField
	err := r.db.From("custom_fields").
		Select("*").
		Eq("entity_type", entityType).
		Eq("entity_id", entityID).
		Execute(&fields)
	if err != nil {
		return nil, err
	}
	if fields == nil {
		fields = []CustomField{}
	}
	return fields, nil
}

func (r *supabaseRepository) CreateCustomField(_ context.Context, entityType, entityID string, data CreateCustomFieldRequest) (*CustomField, error) {
	insertData := map[string]any{
		"entity_type": entityType,
		"entity_id":   entityID,
		"key":         data.Key,
		"value":       data.Value,
	}
	var created []CustomField
	if err := r.db.From("custom_fields").Insert(insertData).Execute(&created); err != nil {
		return nil, err
	}
	return &created[0], nil
}

func (r *supabaseRepository) DeleteCustomField(_ context.Context, fieldID string) error {
	return r.db.From("custom_fields").Eq("id", fieldID).Delete().Execute(nil)
}
