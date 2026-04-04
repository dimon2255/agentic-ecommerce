package assistant

import (
	"encoding/json"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"
)

// Tool input structs for JSON unmarshalling.

type SearchProductsInput struct {
	Query    string `json:"query"`
	Category string `json:"category,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type GetProductDetailsInput struct {
	Slug string `json:"slug"`
}

type GetCategoriesInput struct {
	ParentID string `json:"parent_id,omitempty"`
}

type AddToCartInput struct {
	SKUID    string `json:"sku_id"`
	Quantity int    `json:"quantity"`
}

// GuestTools returns the subset of tools available to unauthenticated users.
// Guests can browse products but cannot access cart functionality.
func GuestTools() []anthropic.Tool {
	all := AllTools()
	guest := make([]anthropic.Tool, 0, 3)
	for _, t := range all {
		if t.Name == "search_products" || t.Name == "get_product_details" || t.Name == "get_categories" {
			guest = append(guest, t)
		}
	}
	return guest
}

// AllTools returns the tool definitions for the AI assistant.
func AllTools() []anthropic.Tool {
	return []anthropic.Tool{
		{
			Name:        "search_products",
			Description: "Search for products in the catalog by text query. Supports optional category filtering and result limit. Use this to find products matching customer needs.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "Search query (e.g. 'wireless headphones', 'laptop under 1500')"
					},
					"category": {
						"type": "string",
						"description": "Optional category slug to filter results (e.g. 'electronics', 'audio')"
					},
					"limit": {
						"type": "integer",
						"description": "Maximum number of results to return (1-20, default 10)",
						"minimum": 1,
						"maximum": 20
					}
				},
				"required": ["query"]
			}`),
		},
		{
			Name:        "get_product_details",
			Description: "Get full details for a specific product including all SKU variants with prices and attributes. Use this when the customer wants to know more about a specific product.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"slug": {
						"type": "string",
						"description": "The product slug (URL-friendly identifier)"
					}
				},
				"required": ["slug"]
			}`),
		},
		{
			Name:        "get_categories",
			Description: "List product categories. Without parent_id returns root categories. With parent_id returns subcategories of that category.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"parent_id": {
						"type": "string",
						"description": "Optional parent category ID to list subcategories. Omit for root categories."
					}
				}
			}`),
		},
		{
			Name:        "get_cart",
			Description: "Get the current contents of the customer's shopping cart including all items, quantities, and prices.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		{
			Name:        "add_to_cart",
			Description: "Add a product SKU to the customer's shopping cart. Always confirm with the customer before calling this tool.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"sku_id": {
						"type": "string",
						"description": "The UUID id of the SKU to add (from the 'id' field in the skus array returned by get_product_details, NOT the sku_code)"
					},
					"quantity": {
						"type": "integer",
						"description": "Number of items to add (default 1)",
						"minimum": 1
					}
				},
				"required": ["sku_id"]
			}`),
		},
	}
}
