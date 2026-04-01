package cart

import "context"

// Repository defines data access operations for the cart domain.
type Repository interface {
	// FindActiveCart finds an active cart by userID (if non-empty) or sessionID.
	FindActiveCart(ctx context.Context, userID, sessionID string) (*Cart, error)

	// CreateCart creates a new active cart with the given session and optional user.
	CreateCart(ctx context.Context, userID, sessionID string) (*Cart, error)

	// GetCartWithItems returns a cart response with enriched items (embedded SKU/product data).
	GetCartWithItems(ctx context.Context, cartID string) (*CartResponse, error)

	// LookupSKUPrice returns the effective price for a SKU (price_override or base_price).
	LookupSKUPrice(ctx context.Context, skuID string) (float64, error)

	// FindCartItem finds a cart item by cart and SKU. Returns nil, nil if not found.
	FindCartItem(ctx context.Context, cartID, skuID string) (*CartItem, error)

	// InsertCartItem adds a new item to the cart.
	InsertCartItem(ctx context.Context, cartID, skuID string, quantity int, unitPrice float64) error

	// UpdateCartItemQuantity sets the quantity on an existing cart item.
	UpdateCartItemQuantity(ctx context.Context, itemID string, quantity int) error

	// VerifyCartItem checks that an item belongs to a cart. Returns nil, nil if not found.
	VerifyCartItem(ctx context.Context, itemID, cartID string) (*CartItem, error)

	// DeleteCartItem removes an item from a cart.
	DeleteCartItem(ctx context.Context, itemID, cartID string) error

	// FindGuestCart finds an active cart for a guest session (user_id is null).
	FindGuestCart(ctx context.Context, sessionID string) (*Cart, error)

	// FindUserCart finds an active cart for a specific user.
	FindUserCart(ctx context.Context, userID string) (*Cart, error)

	// GetCartItems returns all items in a cart.
	GetCartItems(ctx context.Context, cartID string) ([]CartItem, error)

	// MoveCartItem reassigns a cart item to a different cart.
	MoveCartItem(ctx context.Context, itemID, targetCartID string) error

	// UpdateCartStatus sets the status on a cart (e.g., "active", "merged").
	UpdateCartStatus(ctx context.Context, cartID, status string) error
}
