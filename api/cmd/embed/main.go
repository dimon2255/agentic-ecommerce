package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dimon2255/agentic-ecommerce/api/internal/assistant"
	"github.com/dimon2255/agentic-ecommerce/api/internal/config"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/voyage"
)

// Product/SKU types for fetching from Supabase (subset of catalog models).
type product struct {
	ID          string   `json:"id"`
	CategoryID  string   `json:"category_id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	BasePrice   float64  `json:"base_price"`
	Status      string   `json:"status"`
	SKUs        []sku    `json:"skus"`
}

type sku struct {
	ID              string           `json:"id"`
	SKUCode         string           `json:"sku_code"`
	PriceOverride   *float64         `json:"price_override"`
	AttributeValues []attributeValue `json:"sku_attribute_values"`
}

type attributeValue struct {
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
}

type category struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

type customField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type categoryAttribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Assistant.VoyageAPIKey == "" {
		log.Fatal("ESHOP_ASSISTANT_VOYAGE_API_KEY is required")
	}

	db := supabase.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey, cfg.Supabase.Timeout)
	voyageClient := voyage.NewClient(cfg.Assistant.VoyageAPIKey, cfg.Assistant.EmbeddingModel)
	repo := assistant.NewSupabaseRepository(db)
	ctx := context.Background()

	// 1. Fetch all categories (for name lookup)
	var categories []category
	if err := db.From("categories").Select("id,name,parent_id").Execute(&categories); err != nil {
		log.Fatalf("fetch categories: %v", err)
	}
	catMap := make(map[string]category, len(categories))
	for _, c := range categories {
		catMap[c.ID] = c
	}

	// 2. Fetch all category attributes (for name lookup)
	var attrs []categoryAttribute
	if err := db.From("category_attributes").Select("id,name").Execute(&attrs); err != nil {
		log.Fatalf("fetch category attributes: %v", err)
	}
	attrMap := make(map[string]string, len(attrs))
	for _, a := range attrs {
		attrMap[a.ID] = a.Name
	}

	// 3. Fetch all active products with embedded SKUs and attribute values
	var products []product
	err = db.From("products").
		Select("id,category_id,name,description,base_price,status,skus(id,sku_code,price_override,sku_attribute_values(category_attribute_id,value))").
		Eq("status", "active").
		Execute(&products)
	if err != nil {
		log.Fatalf("fetch products: %v", err)
	}

	fmt.Printf("Found %d active products\n", len(products))

	// 4. Fetch all custom fields for products
	var fields []struct {
		EntityID string `json:"entity_id"`
		Key      string `json:"key"`
		Value    string `json:"value"`
	}
	err = db.From("custom_fields").
		Select("entity_id,key,value").
		Eq("entity_type", "product").
		Execute(&fields)
	if err != nil {
		log.Fatalf("fetch custom fields: %v", err)
	}
	cfMap := make(map[string]map[string]string)
	for _, f := range fields {
		if cfMap[f.EntityID] == nil {
			cfMap[f.EntityID] = make(map[string]string)
		}
		cfMap[f.EntityID][f.Key] = f.Value
	}

	// 5. Build embedding content for each product
	texts := make([]string, len(products))
	embProducts := make([]assistant.EmbeddingProduct, len(products))

	for i, p := range products {
		cat := catMap[p.CategoryID]
		parentName := ""
		if cat.ParentID != nil {
			if parent, ok := catMap[*cat.ParentID]; ok {
				parentName = parent.Name
			}
		}

		skus := make([]assistant.EmbeddingSKU, len(p.SKUs))
		for j, s := range p.SKUs {
			attrs := make(map[string]string)
			for _, av := range s.AttributeValues {
				if name, ok := attrMap[av.CategoryAttributeID]; ok {
					attrs[name] = av.Value
				}
			}
			skus[j] = assistant.EmbeddingSKU{
				SKUCode:       s.SKUCode,
				PriceOverride: s.PriceOverride,
				Attributes:    attrs,
			}
		}

		desc := ""
		if p.Description != nil {
			desc = *p.Description
		}

		ep := assistant.EmbeddingProduct{
			ID:             p.ID,
			Name:           p.Name,
			Description:    desc,
			BasePrice:      p.BasePrice,
			Category:       cat.Name,
			ParentCategory: parentName,
			SKUs:           skus,
			CustomFields:   cfMap[p.ID],
		}
		embProducts[i] = ep
		texts[i] = assistant.BuildEmbeddingContent(ep)
	}

	// 6. Batch embed via Voyage AI
	fmt.Println("Generating embeddings via Voyage AI...")
	embeddings, err := voyageClient.Embed(ctx, texts)
	if err != nil {
		log.Fatalf("embed products: %v", err)
	}
	if len(embeddings) != len(products) {
		log.Fatalf("expected %d embeddings, got %d", len(products), len(embeddings))
	}

	// 7. Upsert into product_embeddings
	for i, p := range products {
		record := assistant.EmbeddingRecord{
			ProductID: p.ID,
			Content:   texts[i],
			Embedding: assistant.Float32SliceToVectorString(embeddings[i]),
			Metadata:  assistant.BuildEmbeddingMetadata(embProducts[i]),
		}
		if err := repo.UpsertEmbedding(ctx, record); err != nil {
			log.Printf("  FAILED %s: %v", p.Name, err)
			continue
		}
		fmt.Printf("  Embedded: %s (%d dims)\n", p.Name, len(embeddings[i]))
	}

	fmt.Printf("\nDone! Embedded %d products.\n", len(products))
}
