package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

// mockCatalogService is a minimal mock for catalog.Service used in admin handler tests.
type mockCatalogService struct {
	createProductFn   func(ctx context.Context, req catalog.CreateProductRequest) (*catalog.Product, error)
	getProductBySlug  func(ctx context.Context, slug string) (*catalog.Product, error)
	updateProductFn   func(ctx context.Context, slug string, req catalog.UpdateProductRequest) (*catalog.Product, error)
	deleteProductFn   func(ctx context.Context, slug string) error
	listProductsFn    func(ctx context.Context, filter catalog.ProductFilter) ([]catalog.Product, int, error)
}

func (m *mockCatalogService) ListProducts(ctx context.Context, filter catalog.ProductFilter) ([]catalog.Product, int, error) {
	if m.listProductsFn != nil {
		return m.listProductsFn(ctx, filter)
	}
	return nil, 0, nil
}
func (m *mockCatalogService) GetProductBySlug(ctx context.Context, slug string) (*catalog.Product, error) {
	if m.getProductBySlug != nil {
		return m.getProductBySlug(ctx, slug)
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockCatalogService) CreateProduct(ctx context.Context, req catalog.CreateProductRequest) (*catalog.Product, error) {
	if m.createProductFn != nil {
		return m.createProductFn(ctx, req)
	}
	return nil, nil
}
func (m *mockCatalogService) UpdateProduct(ctx context.Context, slug string, req catalog.UpdateProductRequest) (*catalog.Product, error) {
	if m.updateProductFn != nil {
		return m.updateProductFn(ctx, slug, req)
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockCatalogService) DeleteProduct(ctx context.Context, slug string) error {
	if m.deleteProductFn != nil {
		return m.deleteProductFn(ctx, slug)
	}
	return nil
}

// Stub remaining interface methods
func (m *mockCatalogService) ListCategories(context.Context, catalog.CategoryFilter) ([]catalog.Category, int, error) {
	return nil, 0, nil
}
func (m *mockCatalogService) GetCategoryBySlug(context.Context, string) (*catalog.Category, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockCatalogService) CreateCategory(context.Context, catalog.CreateCategoryRequest) (*catalog.Category, error) {
	return nil, nil
}
func (m *mockCatalogService) UpdateCategory(context.Context, string, catalog.UpdateCategoryRequest) (*catalog.Category, error) {
	return nil, nil
}
func (m *mockCatalogService) DeleteCategory(context.Context, string) error { return nil }
func (m *mockCatalogService) ListAttributesWithOptions(context.Context, string) ([]catalog.CategoryAttribute, error) {
	return nil, nil
}
func (m *mockCatalogService) CreateAttribute(context.Context, string, catalog.CreateAttributeRequest) (*catalog.CategoryAttribute, error) {
	return nil, nil
}
func (m *mockCatalogService) DeleteAttribute(context.Context, string) error { return nil }
func (m *mockCatalogService) ListOptions(context.Context, string) ([]catalog.AttributeOption, error) {
	return nil, nil
}
func (m *mockCatalogService) CreateOption(context.Context, string, catalog.CreateAttributeOptionRequest) (*catalog.AttributeOption, error) {
	return nil, nil
}
func (m *mockCatalogService) DeleteOption(context.Context, string) error { return nil }
func (m *mockCatalogService) ListSKUsWithAttributes(context.Context, string) ([]catalog.SKU, error) {
	return nil, nil
}
func (m *mockCatalogService) CreateSKUWithAttributes(context.Context, string, catalog.CreateSKURequest) (*catalog.SKU, error) {
	return nil, nil
}
func (m *mockCatalogService) DeleteSKU(context.Context, string) error { return nil }
func (m *mockCatalogService) ListCustomFields(context.Context, string, string) ([]catalog.CustomField, error) {
	return nil, nil
}
func (m *mockCatalogService) CreateCustomField(context.Context, string, string, catalog.CreateCustomFieldRequest) (*catalog.CustomField, error) {
	return nil, nil
}
func (m *mockCatalogService) DeleteCustomField(context.Context, string) error { return nil }

// --- Tests ---

func newAuditServiceForTest(t *testing.T) *AuditService {
	t.Helper()
	// Audit service backed by a no-op mock server (fire-and-forget)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("[]"))
	}))
	t.Cleanup(ts.Close)
	db := supabase.NewClient(ts.URL, "test-key", 5*time.Second)
	return NewAuditService(db)
}

func TestCatalogHandler_CreateProduct_HappyPath(t *testing.T) {
	product := &catalog.Product{
		ID:        "prod-1",
		Name:      "Test Product",
		Slug:      "test-product",
		BasePrice: 29.99,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	mock := &mockCatalogService{
		createProductFn: func(_ context.Context, req catalog.CreateProductRequest) (*catalog.Product, error) {
			return product, nil
		},
	}

	handler := NewCatalogHandler(mock, newAuditServiceForTest(t))

	body, _ := json.Marshal(catalog.CreateProductRequest{
		CategoryID: "cat-1",
		Name:       "Test Product",
		Slug:       "test-product",
		BasePrice:  29.99,
		Status:     "active",
	})
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "admin-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateProduct(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp catalog.Product
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID != "prod-1" {
		t.Errorf("expected product ID prod-1, got %s", resp.ID)
	}
}

func TestCatalogHandler_CreateProduct_InvalidBody(t *testing.T) {
	handler := NewCatalogHandler(&mockCatalogService{}, newAuditServiceForTest(t))

	req := httptest.NewRequest("POST", "/products", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	handler.CreateProduct(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCatalogHandler_GetProduct_NotFound(t *testing.T) {
	mock := &mockCatalogService{
		getProductBySlug: func(_ context.Context, slug string) (*catalog.Product, error) {
			return nil, fmt.Errorf("not found")
		},
	}

	handler := NewCatalogHandler(mock, newAuditServiceForTest(t))

	req := httptest.NewRequest("GET", "/products/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.GetProduct(rec, req)

	if rec.Code == http.StatusOK {
		t.Error("expected error status for non-existent product, got 200")
	}
}

func TestCatalogHandler_DeleteProduct_HappyPath(t *testing.T) {
	product := &catalog.Product{ID: "prod-1", Slug: "test-product"}
	mock := &mockCatalogService{
		getProductBySlug: func(_ context.Context, slug string) (*catalog.Product, error) {
			return product, nil
		},
		deleteProductFn: func(_ context.Context, slug string) error {
			return nil
		},
	}

	handler := NewCatalogHandler(mock, newAuditServiceForTest(t))

	req := httptest.NewRequest("DELETE", "/products/test-product", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "admin-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "test-product")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.DeleteProduct(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}
