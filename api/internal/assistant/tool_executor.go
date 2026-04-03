package assistant

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/pagination"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"
)

const maxToolResultBytes = 4000

// ToolExecutor dispatches Claude tool calls to the appropriate service methods.
type ToolExecutor struct {
	catalogSvc catalog.Service
	cartSvc    cart.Service
}

// NewToolExecutor creates a tool executor wired to the catalog and cart services.
func NewToolExecutor(catalogSvc catalog.Service, cartSvc cart.Service) *ToolExecutor {
	return &ToolExecutor{catalogSvc: catalogSvc, cartSvc: cartSvc}
}

// ExecuteResult holds the outcome of a tool execution.
type ExecuteResult struct {
	Content     string
	IsError     bool
	CartUpdated bool
}

// Execute runs a tool call and returns a JSON result string for Claude.
func (te *ToolExecutor) Execute(ctx context.Context, block anthropic.ContentBlock, userID string) ExecuteResult {
	switch block.Name {
	case "search_products":
		return te.searchProducts(ctx, block.Input)
	case "get_product_details":
		return te.getProductDetails(ctx, block.Input)
	case "get_categories":
		return te.getCategories(ctx, block.Input)
	case "get_cart":
		return te.getCart(ctx, userID)
	case "add_to_cart":
		return te.addToCart(ctx, block.Input, userID)
	default:
		return ExecuteResult{Content: fmt.Sprintf("Unknown tool: %s", block.Name), IsError: true}
	}
}

func (te *ToolExecutor) searchProducts(ctx context.Context, raw json.RawMessage) ExecuteResult {
	var input SearchProductsInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return ExecuteResult{Content: "Invalid input: " + err.Error(), IsError: true}
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	filter := catalog.ProductFilter{
		Search: input.Query,
		Params: pagination.Params{Page: 1, PerPage: limit},
	}
	if input.Category != "" {
		filter.CategoryID = &input.Category
	}

	products, total, err := te.catalogSvc.ListProducts(ctx, filter)
	if err != nil {
		return ExecuteResult{Content: "Failed to search products: " + err.Error(), IsError: true}
	}

	type result struct {
		Products []catalog.Product `json:"products"`
		Total    int               `json:"total"`
	}
	return marshalResult(result{Products: products, Total: total})
}

func (te *ToolExecutor) getProductDetails(ctx context.Context, raw json.RawMessage) ExecuteResult {
	var input GetProductDetailsInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return ExecuteResult{Content: "Invalid input: " + err.Error(), IsError: true}
	}

	product, err := te.catalogSvc.GetProductBySlug(ctx, input.Slug)
	if err != nil {
		return ExecuteResult{Content: "Product not found: " + input.Slug, IsError: true}
	}

	skus, err := te.catalogSvc.ListSKUsWithAttributes(ctx, product.ID)
	if err != nil {
		return ExecuteResult{Content: "Failed to fetch SKUs: " + err.Error(), IsError: true}
	}

	type result struct {
		Product catalog.Product `json:"product"`
		SKUs    []catalog.SKU   `json:"skus"`
	}
	return marshalResult(result{Product: *product, SKUs: skus})
}

func (te *ToolExecutor) getCategories(ctx context.Context, raw json.RawMessage) ExecuteResult {
	var input GetCategoriesInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return ExecuteResult{Content: "Invalid input: " + err.Error(), IsError: true}
	}

	filter := catalog.CategoryFilter{
		Params: pagination.Params{Page: 1, PerPage: 50},
	}
	if input.ParentID != "" {
		filter.ParentID = &input.ParentID
	}

	categories, _, err := te.catalogSvc.ListCategories(ctx, filter)
	if err != nil {
		return ExecuteResult{Content: "Failed to list categories: " + err.Error(), IsError: true}
	}

	type result struct {
		Categories []catalog.Category `json:"categories"`
	}
	return marshalResult(result{Categories: categories})
}

func (te *ToolExecutor) getCart(ctx context.Context, userID string) ExecuteResult {
	cartResp, err := te.cartSvc.GetCart(ctx, userID, "")
	if err != nil {
		return ExecuteResult{Content: "Failed to get cart: " + err.Error(), IsError: true}
	}
	return marshalResult(cartResp)
}

func (te *ToolExecutor) addToCart(ctx context.Context, raw json.RawMessage, userID string) ExecuteResult {
	var input AddToCartInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return ExecuteResult{Content: "Invalid input: " + err.Error(), IsError: true}
	}

	qty := input.Quantity
	if qty <= 0 {
		qty = 1
	}

	cartResp, err := te.cartSvc.AddItem(ctx, userID, "", cart.AddItemRequest{
		SKUID:    input.SKUID,
		Quantity: qty,
	})
	if err != nil {
		return ExecuteResult{Content: "Failed to add to cart: " + err.Error(), IsError: true}
	}

	return ExecuteResult{Content: marshalJSON(cartResp), CartUpdated: true}
}

// marshalResult serializes data to JSON, truncating if over maxToolResultBytes.
func marshalResult(v any) ExecuteResult {
	return ExecuteResult{Content: marshalJSON(v)}
}

func marshalJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return `{"error":"failed to serialize result"}`
	}
	s := string(data)
	if len(s) > maxToolResultBytes {
		s = s[:maxToolResultBytes] + "...(truncated)"
	}
	return s
}
