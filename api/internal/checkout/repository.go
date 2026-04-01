package checkout

import "context"

// Repository defines data access operations for the checkout domain.
type Repository interface {
	FindActiveCart(ctx context.Context, userID, sessionID string) (*cart, error)
	GetCartItems(ctx context.Context, cartID string) ([]cartItem, error)
	UpdateCartItemPrice(ctx context.Context, itemID string, price float64) error
	CreateOrder(ctx context.Context, data map[string]any) (*Order, error)
	CreateOrderItem(ctx context.Context, data map[string]any) error
	DeleteOrder(ctx context.Context, orderID string) error
	UpdateOrder(ctx context.Context, orderID string, data map[string]any) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	GetOrderItems(ctx context.Context, orderID string) ([]OrderItem, error)
	FindOrderByPaymentIntent(ctx context.Context, piID string) (*Order, error)
	UpdateOrderByPaymentIntent(ctx context.Context, piID string, data map[string]any) error
	ExpireUserCarts(ctx context.Context, userID string) error
}
