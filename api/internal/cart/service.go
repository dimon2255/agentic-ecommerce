package cart

import "context"

// Service defines business operations for the cart domain.
type Service interface {
	// GetCart returns the active cart for the given user or session.
	GetCart(ctx context.Context, userID, sessionID string) (*CartResponse, error)

	// AddItem adds an item to the cart, creating the cart if needed.
	AddItem(ctx context.Context, userID, sessionID string, req AddItemRequest) (*CartResponse, error)

	// UpdateItem changes the quantity of an existing cart item.
	UpdateItem(ctx context.Context, userID, sessionID, itemID string, req UpdateItemRequest) (*CartResponse, error)

	// RemoveItem deletes an item from the cart.
	RemoveItem(ctx context.Context, userID, sessionID, itemID string) error

	// MergeCart merges a guest session cart into the authenticated user's cart.
	MergeCart(ctx context.Context, userID string, req MergeCartRequest) (*CartResponse, error)
}
