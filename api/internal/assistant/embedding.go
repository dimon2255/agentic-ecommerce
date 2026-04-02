package assistant

import (
	"fmt"
	"strings"
)

// EmbeddingProduct holds the denormalized product data needed to build embedding content.
type EmbeddingProduct struct {
	ID          string
	Name        string
	Description string
	BasePrice   float64
	Category    string // e.g. "Laptops"
	ParentCategory string // e.g. "Electronics" (empty if no parent)
	SKUs        []EmbeddingSKU
	CustomFields map[string]string // key → value
}

// EmbeddingSKU holds SKU data for embedding content.
type EmbeddingSKU struct {
	SKUCode       string
	PriceOverride *float64
	Attributes    map[string]string // attribute name → value
}

// BuildEmbeddingContent constructs a rich text representation of a product for embedding.
// The output is natural language optimized for semantic search retrieval.
func BuildEmbeddingContent(p EmbeddingProduct) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Product: %s\n", p.Name))

	// Category path
	if p.ParentCategory != "" {
		sb.WriteString(fmt.Sprintf("Category: %s > %s\n", p.ParentCategory, p.Category))
	} else {
		sb.WriteString(fmt.Sprintf("Category: %s\n", p.Category))
	}

	if p.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}

	sb.WriteString(fmt.Sprintf("Base Price: $%.2f\n", p.BasePrice))

	// SKU variants
	if len(p.SKUs) > 0 {
		sb.WriteString("\nAvailable Variants:\n")
		for _, sku := range p.SKUs {
			price := p.BasePrice
			if sku.PriceOverride != nil {
				price = *sku.PriceOverride
			}

			attrs := formatAttributes(sku.Attributes)
			if attrs != "" {
				sb.WriteString(fmt.Sprintf("- %s: %s — $%.2f\n", sku.SKUCode, attrs, price))
			} else {
				sb.WriteString(fmt.Sprintf("- %s — $%.2f\n", sku.SKUCode, price))
			}
		}
	}

	// Custom fields (tags, supplier, etc.)
	if len(p.CustomFields) > 0 {
		sb.WriteString("\n")
		for key, value := range p.CustomFields {
			sb.WriteString(fmt.Sprintf("%s: %s\n", capitalizeFirst(key), value))
		}
	}

	return strings.TrimSpace(sb.String())
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatAttributes(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(attrs))
	for name, value := range attrs {
		parts = append(parts, fmt.Sprintf("%s %s", value, name))
	}
	return strings.Join(parts, ", ")
}

// BuildEmbeddingMetadata creates the JSONB metadata stored alongside the embedding.
func BuildEmbeddingMetadata(p EmbeddingProduct) map[string]any {
	return map[string]any{
		"product_name": p.Name,
		"category":     p.Category,
		"base_price":   p.BasePrice,
		"sku_count":    len(p.SKUs),
	}
}
