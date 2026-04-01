package cart

import (
	"context"
	"fmt"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type supabaseRepository struct {
	db *supabase.Client
}

// NewSupabaseRepository creates a cart repository backed by Supabase PostgREST.
func NewSupabaseRepository(db *supabase.Client) Repository {
	return &supabaseRepository{db: db}
}

func (r *supabaseRepository) FindActiveCart(_ context.Context, userID, sessionID string) (*Cart, error) {
	q := r.db.From("carts").Select("*").Eq("status", "active")
	if userID != "" {
		q = q.Eq("user_id", userID)
	} else if sessionID != "" {
		q = q.Eq("session_id", sessionID).Is("user_id", "null")
	} else {
		return nil, nil
	}

	var carts []Cart
	if err := q.Limit(1).Execute(&carts); err != nil {
		return nil, fmt.Errorf("find active cart: %w", err)
	}
	if len(carts) == 0 {
		return nil, nil
	}
	return &carts[0], nil
}

func (r *supabaseRepository) CreateCart(_ context.Context, userID, sessionID string) (*Cart, error) {
	data := map[string]any{
		"session_id": sessionID,
		"status":     "active",
	}
	if userID != "" {
		data["user_id"] = userID
	}

	var created []Cart
	if err := r.db.From("carts").Insert(data).Execute(&created); err != nil {
		return nil, fmt.Errorf("create cart: %w", err)
	}
	if len(created) == 0 {
		return nil, fmt.Errorf("cart not returned after creation")
	}
	return &created[0], nil
}

func (r *supabaseRepository) GetCartWithItems(_ context.Context, cartID string) (*CartResponse, error) {
	var items []CartItemWithSKU
	err := r.db.From("cart_items").
		Select("*,skus(sku_code,price_override,products(name,slug,base_price,images))").
		Eq("cart_id", cartID).
		Execute(&items)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	if items == nil {
		items = []CartItemWithSKU{}
	}
	return &CartResponse{ID: cartID, Items: items}, nil
}

func (r *supabaseRepository) LookupSKUPrice(_ context.Context, skuID string) (float64, error) {
	var skus []SKUForPrice
	err := r.db.From("skus").
		Select("price_override,products(base_price)").
		Eq("id", skuID).
		Execute(&skus)
	if err != nil {
		return 0, fmt.Errorf("lookup sku price: %w", err)
	}
	if len(skus) == 0 {
		return 0, fmt.Errorf("SKU not found")
	}
	sku := skus[0]
	if sku.PriceOverride != nil {
		return *sku.PriceOverride, nil
	}
	return sku.Product.BasePrice, nil
}

func (r *supabaseRepository) FindCartItem(_ context.Context, cartID, skuID string) (*CartItem, error) {
	var items []CartItem
	err := r.db.From("cart_items").Select("*").
		Eq("cart_id", cartID).Eq("sku_id", skuID).
		Execute(&items)
	if err != nil {
		return nil, fmt.Errorf("find cart item: %w", err)
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *supabaseRepository) InsertCartItem(_ context.Context, cartID, skuID string, quantity int, unitPrice float64) error {
	var inserted []CartItem
	err := r.db.From("cart_items").Insert(map[string]any{
		"cart_id":    cartID,
		"sku_id":     skuID,
		"quantity":   quantity,
		"unit_price": unitPrice,
	}).Execute(&inserted)
	if err != nil {
		return fmt.Errorf("insert cart item: %w", err)
	}
	return nil
}

func (r *supabaseRepository) UpdateCartItemQuantity(_ context.Context, itemID string, quantity int) error {
	var updated []CartItem
	err := r.db.From("cart_items").
		Update(map[string]any{"quantity": quantity}).
		Eq("id", itemID).
		Execute(&updated)
	if err != nil {
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	return nil
}

func (r *supabaseRepository) VerifyCartItem(_ context.Context, itemID, cartID string) (*CartItem, error) {
	var items []CartItem
	err := r.db.From("cart_items").Select("*").
		Eq("id", itemID).Eq("cart_id", cartID).
		Execute(&items)
	if err != nil {
		return nil, fmt.Errorf("verify cart item: %w", err)
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *supabaseRepository) DeleteCartItem(_ context.Context, itemID, cartID string) error {
	err := r.db.From("cart_items").Delete().
		Eq("id", itemID).
		Eq("cart_id", cartID).
		Execute(nil)
	if err != nil {
		return fmt.Errorf("delete cart item: %w", err)
	}
	return nil
}

func (r *supabaseRepository) FindGuestCart(_ context.Context, sessionID string) (*Cart, error) {
	var carts []Cart
	err := r.db.From("carts").Select("*").
		Eq("session_id", sessionID).
		Is("user_id", "null").
		Eq("status", "active").
		Execute(&carts)
	if err != nil {
		return nil, fmt.Errorf("find guest cart: %w", err)
	}
	if len(carts) == 0 {
		return nil, nil
	}
	return &carts[0], nil
}

func (r *supabaseRepository) FindUserCart(_ context.Context, userID string) (*Cart, error) {
	var carts []Cart
	err := r.db.From("carts").Select("*").
		Eq("user_id", userID).
		Eq("status", "active").
		Limit(1).
		Execute(&carts)
	if err != nil {
		return nil, fmt.Errorf("find user cart: %w", err)
	}
	if len(carts) == 0 {
		return nil, nil
	}
	return &carts[0], nil
}

func (r *supabaseRepository) GetCartItems(_ context.Context, cartID string) ([]CartItem, error) {
	var items []CartItem
	err := r.db.From("cart_items").Select("*").
		Eq("cart_id", cartID).
		Execute(&items)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	if items == nil {
		items = []CartItem{}
	}
	return items, nil
}

func (r *supabaseRepository) MoveCartItem(_ context.Context, itemID, targetCartID string) error {
	err := r.db.From("cart_items").
		Update(map[string]any{"cart_id": targetCartID}).
		Eq("id", itemID).
		Execute(nil)
	if err != nil {
		return fmt.Errorf("move cart item: %w", err)
	}
	return nil
}

func (r *supabaseRepository) UpdateCartStatus(_ context.Context, cartID, status string) error {
	err := r.db.From("carts").
		Update(map[string]any{"status": status}).
		Eq("id", cartID).
		Execute(nil)
	if err != nil {
		return fmt.Errorf("update cart status: %w", err)
	}
	return nil
}
