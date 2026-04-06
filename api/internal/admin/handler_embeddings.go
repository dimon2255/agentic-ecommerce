package admin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/assistant"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/voyage"
)

// Supabase DTO types for embedding data fetches.
type embProduct struct {
	ID          string   `json:"id"`
	CategoryID  string   `json:"category_id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	BasePrice   float64  `json:"base_price"`
	Status      string   `json:"status"`
	SKUs        []embSKU `json:"skus"`
}

type embSKU struct {
	ID              string         `json:"id"`
	SKUCode         string         `json:"sku_code"`
	PriceOverride   *float64       `json:"price_override"`
	AttributeValues []embAttrValue `json:"sku_attribute_values"`
}

type embAttrValue struct {
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
}

type embCategory struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

type embCategoryAttribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EmbeddingHandler manages product embedding generation via the admin API.
type EmbeddingHandler struct {
	db     *supabase.Client
	voyage *voyage.Client
	repo   assistant.Repository
	audit  *AuditService
}

func NewEmbeddingHandler(db *supabase.Client, v *voyage.Client, repo assistant.Repository, audit *AuditService) *EmbeddingHandler {
	return &EmbeddingHandler{db: db, voyage: v, repo: repo, audit: audit}
}

func (h *EmbeddingHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/regenerate", h.RegenerateAll)
	r.Post("/regenerate/{productId}", h.RegenerateProduct)
	r.Get("/status", h.Status)
	return r
}

// RegenerateAll re-embeds all active products in a background goroutine.
func (h *EmbeddingHandler) RegenerateAll(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	go func() {
		ctx := context.Background()
		count, err := h.regenerateAll(ctx)
		if err != nil {
			slog.Error("embedding regeneration failed", "error", err)
			return
		}
		slog.Info("embedding regeneration complete", "count", count, "triggered_by", userID)
	}()

	h.audit.LogFromRequest(r, userID, "embeddings:regenerate_all", "embeddings", "", nil)
	response.JSON(w, http.StatusAccepted, map[string]string{
		"status":  "started",
		"message": "Regenerating all product embeddings in background",
	})
}

// RegenerateProduct re-embeds a single product synchronously.
func (h *EmbeddingHandler) RegenerateProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	userID, _ := middleware.GetUserID(r.Context())

	if err := h.RegenerateProductByID(r.Context(), productID); err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("embedding failed: %v", err))
		return
	}

	h.audit.LogFromRequest(r, userID, "embeddings:regenerate", "product", productID, nil)
	response.JSON(w, http.StatusOK, map[string]string{
		"status":     "completed",
		"product_id": productID,
	})
}

// Status returns embedding coverage counts.
func (h *EmbeddingHandler) Status(w http.ResponseWriter, r *http.Request) {
	var products []struct {
		ID string `json:"id"`
	}
	h.db.From("products").Select("id").Eq("status", "active").Execute(&products)

	var embeddings []struct {
		ProductID string `json:"product_id"`
	}
	h.db.From("product_embeddings").Select("product_id").Execute(&embeddings)

	response.JSON(w, http.StatusOK, map[string]any{
		"active_products":   len(products),
		"embedded_products": len(embeddings),
		"coverage_complete": len(embeddings) >= len(products),
	})
}

// RegenerateProductByID re-embeds a single product. Exported for use as a hook.
func (h *EmbeddingHandler) RegenerateProductByID(ctx context.Context, productID string) error {
	catMap, attrMap, cfMap, err := h.fetchLookupData()
	if err != nil {
		return err
	}

	var products []embProduct
	err = h.db.From("products").
		Select("id,category_id,name,description,base_price,status,skus(id,sku_code,price_override,sku_attribute_values(category_attribute_id,value))").
		Eq("id", productID).
		Execute(&products)
	if err != nil {
		return fmt.Errorf("fetch product: %w", err)
	}
	if len(products) == 0 {
		return fmt.Errorf("product not found: %s", productID)
	}

	p := products[0]
	ep := h.buildEmbeddingProduct(p, catMap, attrMap, cfMap)
	text := assistant.BuildEmbeddingContent(ep)

	embeddings, err := h.voyage.Embed(ctx, []string{text})
	if err != nil {
		return fmt.Errorf("voyage embed: %w", err)
	}

	return h.repo.UpsertEmbedding(ctx, assistant.EmbeddingRecord{
		ProductID: p.ID,
		Content:   text,
		Embedding: assistant.Float32SliceToVectorString(embeddings[0]),
		Metadata:  assistant.BuildEmbeddingMetadata(ep),
	})
}

// regenerateAll re-embeds all active products.
func (h *EmbeddingHandler) regenerateAll(ctx context.Context) (int, error) {
	catMap, attrMap, cfMap, err := h.fetchLookupData()
	if err != nil {
		return 0, err
	}

	var products []embProduct
	err = h.db.From("products").
		Select("id,category_id,name,description,base_price,status,skus(id,sku_code,price_override,sku_attribute_values(category_attribute_id,value))").
		Eq("status", "active").
		Execute(&products)
	if err != nil {
		return 0, fmt.Errorf("fetch products: %w", err)
	}

	if len(products) == 0 {
		return 0, nil
	}

	texts := make([]string, len(products))
	embProducts := make([]assistant.EmbeddingProduct, len(products))
	for i, p := range products {
		ep := h.buildEmbeddingProduct(p, catMap, attrMap, cfMap)
		embProducts[i] = ep
		texts[i] = assistant.BuildEmbeddingContent(ep)
	}

	embeddings, err := h.voyage.Embed(ctx, texts)
	if err != nil {
		return 0, fmt.Errorf("voyage embed: %w", err)
	}

	embedded := 0
	for i, p := range products {
		record := assistant.EmbeddingRecord{
			ProductID: p.ID,
			Content:   texts[i],
			Embedding: assistant.Float32SliceToVectorString(embeddings[i]),
			Metadata:  assistant.BuildEmbeddingMetadata(embProducts[i]),
		}
		if err := h.repo.UpsertEmbedding(ctx, record); err != nil {
			slog.Error("upsert embedding failed", "product_id", p.ID, "error", err)
			continue
		}
		embedded++
	}

	return embedded, nil
}

func (h *EmbeddingHandler) fetchLookupData() (map[string]embCategory, map[string]string, map[string]map[string]string, error) {
	var cats []embCategory
	if err := h.db.From("categories").Select("id,name,parent_id").Execute(&cats); err != nil {
		return nil, nil, nil, fmt.Errorf("fetch categories: %w", err)
	}
	catMap := make(map[string]embCategory, len(cats))
	for _, c := range cats {
		catMap[c.ID] = c
	}

	var attrs []embCategoryAttribute
	if err := h.db.From("category_attributes").Select("id,name").Execute(&attrs); err != nil {
		return nil, nil, nil, fmt.Errorf("fetch attributes: %w", err)
	}
	attrMap := make(map[string]string, len(attrs))
	for _, a := range attrs {
		attrMap[a.ID] = a.Name
	}

	var fields []struct {
		EntityID string `json:"entity_id"`
		Key      string `json:"key"`
		Value    string `json:"value"`
	}
	if err := h.db.From("custom_fields").Select("entity_id,key,value").Eq("entity_type", "product").Execute(&fields); err != nil {
		return nil, nil, nil, fmt.Errorf("fetch custom fields: %w", err)
	}
	cfMap := make(map[string]map[string]string)
	for _, f := range fields {
		if cfMap[f.EntityID] == nil {
			cfMap[f.EntityID] = make(map[string]string)
		}
		cfMap[f.EntityID][f.Key] = f.Value
	}

	return catMap, attrMap, cfMap, nil
}

func (h *EmbeddingHandler) buildEmbeddingProduct(p embProduct, catMap map[string]embCategory, attrMap map[string]string, cfMap map[string]map[string]string) assistant.EmbeddingProduct {
	cat := catMap[p.CategoryID]
	parentName := ""
	if cat.ParentID != nil {
		if parent, ok := catMap[*cat.ParentID]; ok {
			parentName = parent.Name
		}
	}

	skus := make([]assistant.EmbeddingSKU, len(p.SKUs))
	for j, s := range p.SKUs {
		skuAttrs := make(map[string]string)
		for _, av := range s.AttributeValues {
			if name, ok := attrMap[av.CategoryAttributeID]; ok {
				skuAttrs[name] = av.Value
			}
		}
		skus[j] = assistant.EmbeddingSKU{
			SKUCode:       s.SKUCode,
			PriceOverride: s.PriceOverride,
			Attributes:    skuAttrs,
		}
	}

	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}

	return assistant.EmbeddingProduct{
		ID:             p.ID,
		Name:           p.Name,
		Description:    desc,
		BasePrice:      p.BasePrice,
		Category:       cat.Name,
		ParentCategory: parentName,
		SKUs:           skus,
		CustomFields:   cfMap[p.ID],
	}
}
