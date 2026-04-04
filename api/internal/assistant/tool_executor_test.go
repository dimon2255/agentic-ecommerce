package assistant

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"
)

func newTestExecutor() *ToolExecutor {
	catSvc := &mockCatalogService{
		products: []catalog.Product{
			{ID: "p1", Name: "ProBook 15", Slug: "probook-15", BasePrice: 999.99, Status: "active"},
			{ID: "p2", Name: "UltraSound X3", Slug: "ultrasound-x3", BasePrice: 299.99, Status: "active"},
		},
		skus: []catalog.SKU{
			{ID: "s1", ProductID: "p1", SKUCode: "PB15-16-512", Status: "active"},
		},
		categories: []catalog.Category{
			{ID: "c1", Name: "Electronics", Slug: "electronics"},
			{ID: "c2", Name: "Audio", Slug: "audio"},
		},
	}
	cartSvc := &mockCartService{
		cartResp: &cart.CartResponse{
			ID: "cart-1",
			Items: []cart.CartItemWithSKU{
				{ID: "item-1", SKUID: "s1", Quantity: 2, UnitPrice: 999.99},
			},
		},
	}
	return NewToolExecutor(catSvc, cartSvc)
}

func TestExecute_SearchProducts(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "search_products",
		Input: json.RawMessage(`{"query":"laptop","limit":5}`),
	}, "user-1", false)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "ProBook 15") {
		t.Errorf("expected product in result: %s", result.Content)
	}
	if result.CartUpdated {
		t.Error("search should not update cart")
	}
}

func TestExecute_GetProductDetails(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "get_product_details",
		Input: json.RawMessage(`{"slug":"probook-15"}`),
	}, "user-1", false)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "ProBook 15") {
		t.Errorf("expected product details: %s", result.Content)
	}
	if !strings.Contains(result.Content, "PB15-16-512") {
		t.Errorf("expected SKU in details: %s", result.Content)
	}
}

func TestExecute_GetProductDetails_NotFound(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "get_product_details",
		Input: json.RawMessage(`{"slug":"nonexistent"}`),
	}, "user-1", false)

	if !result.IsError {
		t.Error("expected error for nonexistent product")
	}
	if !strings.Contains(result.Content, "not found") {
		t.Errorf("expected 'not found' in error: %s", result.Content)
	}
}

func TestExecute_GetCategories(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "get_categories",
		Input: json.RawMessage(`{}`),
	}, "user-1", false)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Electronics") {
		t.Errorf("expected categories: %s", result.Content)
	}
}

func TestExecute_GetCart(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "get_cart",
		Input: json.RawMessage(`{}`),
	}, "user-1", false)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "cart-1") {
		t.Errorf("expected cart data: %s", result.Content)
	}
	if result.CartUpdated {
		t.Error("get_cart should not set CartUpdated")
	}
}

func TestExecute_AddToCart(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "add_to_cart",
		Input: json.RawMessage(`{"sku_id":"s1","quantity":1}`),
	}, "user-1", false)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if !result.CartUpdated {
		t.Error("add_to_cart should set CartUpdated=true")
	}
}

func TestExecute_UnknownTool(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "hack_the_planet",
		Input: json.RawMessage(`{}`),
	}, "user-1", false)

	if !result.IsError {
		t.Error("expected error for unknown tool")
	}
	if !strings.Contains(result.Content, "Unknown tool") {
		t.Errorf("expected 'Unknown tool' error: %s", result.Content)
	}
}

func TestExecute_InvalidInput(t *testing.T) {
	te := newTestExecutor()
	result := te.Execute(context.Background(), anthropic.ContentBlock{
		Name:  "search_products",
		Input: json.RawMessage(`{invalid json`),
	}, "user-1", false)

	if !result.IsError {
		t.Error("expected error for invalid JSON input")
	}
}
