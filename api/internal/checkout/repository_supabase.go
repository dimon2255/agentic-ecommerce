package checkout

import (
	"context"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type supabaseRepository struct {
	db *supabase.Client
}

func NewSupabaseRepository(db *supabase.Client) Repository {
	return &supabaseRepository{db: db}
}

func (r *supabaseRepository) FindActiveCart(_ context.Context, userID, sessionID string) (*cart, error) {
	if userID != "" {
		var carts []cart
		err := r.db.From("carts").Select("*").Eq("user_id", userID).Eq("status", "active").Limit(1).Execute(&carts)
		if err != nil {
			return nil, err
		}
		if len(carts) > 0 {
			return &carts[0], nil
		}
	}
	if sessionID != "" {
		var carts []cart
		err := r.db.From("carts").Select("*").Eq("session_id", sessionID).Eq("status", "active").Is("user_id", "null").Limit(1).Execute(&carts)
		if err != nil {
			return nil, err
		}
		if len(carts) > 0 {
			return &carts[0], nil
		}
	}
	return nil, nil
}

func (r *supabaseRepository) GetCartItems(_ context.Context, cartID string) ([]cartItem, error) {
	var items []cartItem
	err := r.db.From("cart_items").
		Select("*,skus(sku_code,price_override,products(name,base_price))").
		Eq("cart_id", cartID).
		Execute(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *supabaseRepository) UpdateCartItemPrice(_ context.Context, itemID string, price float64) error {
	return r.db.From("cart_items").Update(map[string]any{"unit_price": price}).Eq("id", itemID).Execute(nil)
}

func (r *supabaseRepository) CreateOrder(_ context.Context, data map[string]any) (*Order, error) {
	var orders []Order
	if err := r.db.From("orders").Insert(data).Execute(&orders); err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}
	return &orders[0], nil
}

func (r *supabaseRepository) CreateOrderItem(_ context.Context, data map[string]any) error {
	return r.db.From("order_items").Insert(data).Execute(nil)
}

func (r *supabaseRepository) DeleteOrder(_ context.Context, orderID string) error {
	return r.db.From("orders").Delete().Eq("id", orderID).Execute(nil)
}

func (r *supabaseRepository) UpdateOrder(_ context.Context, orderID string, data map[string]any) error {
	return r.db.From("orders").Update(data).Eq("id", orderID).Execute(nil)
}

func (r *supabaseRepository) GetOrder(_ context.Context, orderID string) (*Order, error) {
	var orders []Order
	err := r.db.From("orders").Select("*").Eq("id", orderID).Limit(1).Execute(&orders)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}
	return &orders[0], nil
}

func (r *supabaseRepository) GetOrderItems(_ context.Context, orderID string) ([]OrderItem, error) {
	var items []OrderItem
	err := r.db.From("order_items").Select("*").Eq("order_id", orderID).Execute(&items)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []OrderItem{}
	}
	return items, nil
}

func (r *supabaseRepository) FindOrderByPaymentIntent(_ context.Context, piID string) (*Order, error) {
	var orders []Order
	err := r.db.From("orders").Select("user_id").Eq("stripe_payment_intent_id", piID).Execute(&orders)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}
	return &orders[0], nil
}

func (r *supabaseRepository) UpdateOrderByPaymentIntent(_ context.Context, piID string, data map[string]any) error {
	return r.db.From("orders").Update(data).Eq("stripe_payment_intent_id", piID).Execute(nil)
}

func (r *supabaseRepository) ExpireUserCarts(_ context.Context, userID string) error {
	return r.db.From("carts").Update(map[string]any{"status": "expired"}).Eq("user_id", userID).Eq("status", "active").Execute(nil)
}
