package cart

import (
	"context"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
)

type cartService struct {
	repo Repository
}

// NewService creates a cart service backed by the given repository.
func NewService(repo Repository) Service {
	return &cartService{repo: repo}
}

func (s *cartService) GetCart(ctx context.Context, userID, sessionID string) (*CartResponse, error) {
	if userID == "" && sessionID == "" {
		return nil, apperror.NewUnauthorized("authentication or session ID required")
	}

	cart, err := s.repo.FindActiveCart(ctx, userID, sessionID)
	if err != nil {
		return nil, apperror.NewInternal("failed to find cart", err)
	}
	if cart == nil {
		return &CartResponse{Items: []CartItemWithSKU{}}, nil
	}

	resp, err := s.repo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch cart items", err)
	}
	return resp, nil
}

func (s *cartService) AddItem(ctx context.Context, userID, sessionID string, req AddItemRequest) (*CartResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find or create cart
	cart, err := s.findOrCreateCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	// Lookup current SKU price
	unitPrice, err := s.repo.LookupSKUPrice(ctx, req.SKUID)
	if err != nil {
		return nil, apperror.NewInvalidInput("invalid SKU", nil)
	}

	// Atomic upsert via RPC — inserts or increments quantity in a single DB call
	if err := s.repo.InsertCartItem(ctx, cart.ID, req.SKUID, req.Quantity, unitPrice); err != nil {
		return nil, apperror.NewInternal("failed to add item to cart", err)
	}

	resp, err := s.repo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch updated cart", err)
	}
	return resp, nil
}

func (s *cartService) UpdateItem(ctx context.Context, userID, sessionID, itemID string, req UpdateItemRequest) (*CartResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	cart, err := s.repo.FindActiveCart(ctx, userID, sessionID)
	if err != nil {
		return nil, apperror.NewInternal("failed to find cart", err)
	}
	if cart == nil {
		return nil, apperror.NewNotFound("cart")
	}

	// Verify item belongs to this cart
	item, err := s.repo.VerifyCartItem(ctx, itemID, cart.ID)
	if err != nil {
		return nil, apperror.NewInternal("failed to verify cart item", err)
	}
	if item == nil {
		return nil, apperror.NewNotFound("cart item")
	}

	if err := s.repo.UpdateCartItemQuantity(ctx, itemID, req.Quantity); err != nil {
		return nil, apperror.NewInternal("failed to update item", err)
	}

	resp, err := s.repo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch updated cart", err)
	}
	return resp, nil
}

func (s *cartService) RemoveItem(ctx context.Context, userID, sessionID, itemID string) error {
	cart, err := s.repo.FindActiveCart(ctx, userID, sessionID)
	if err != nil {
		return apperror.NewInternal("failed to find cart", err)
	}
	if cart == nil {
		return apperror.NewNotFound("cart")
	}

	// Verify item belongs to this cart
	item, err := s.repo.VerifyCartItem(ctx, itemID, cart.ID)
	if err != nil {
		return apperror.NewInternal("failed to verify cart item", err)
	}
	if item == nil {
		return apperror.NewNotFound("cart item")
	}

	if err := s.repo.DeleteCartItem(ctx, itemID, cart.ID); err != nil {
		return apperror.NewInternal("failed to remove item", err)
	}
	return nil
}

func (s *cartService) MergeCart(ctx context.Context, userID string, req MergeCartRequest) (*CartResponse, error) {
	if userID == "" {
		return nil, apperror.NewUnauthorized("authentication required")
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find guest cart
	guestCart, err := s.repo.FindGuestCart(ctx, req.SessionID)
	if err != nil {
		return nil, apperror.NewInternal("failed to find guest cart", err)
	}

	if guestCart == nil {
		// No guest cart to merge — return current user cart or empty
		userCart, err := s.repo.FindUserCart(ctx, userID)
		if err != nil {
			return nil, apperror.NewInternal("failed to find user cart", err)
		}
		if userCart == nil {
			return &CartResponse{Items: []CartItemWithSKU{}}, nil
		}
		resp, err := s.repo.GetCartWithItems(ctx, userCart.ID)
		if err != nil {
			return nil, apperror.NewInternal("failed to fetch cart", err)
		}
		return resp, nil
	}

	// Find or create user cart
	userCart, err := s.repo.FindUserCart(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("failed to find user cart", err)
	}
	if userCart == nil {
		userCart, err = s.repo.CreateCart(ctx, userID, req.SessionID)
		if err != nil {
			return nil, apperror.NewInternal("failed to create user cart", err)
		}
	}

	// Fetch guest cart items
	guestItems, err := s.repo.GetCartItems(ctx, guestCart.ID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch guest cart items", err)
	}

	// Move each guest item to user cart
	for _, item := range guestItems {
		// Check for duplicate SKU in user cart
		existing, err := s.repo.FindCartItem(ctx, userCart.ID, item.SKUID)
		if err != nil {
			return nil, apperror.NewInternal("failed to check duplicate items", err)
		}

		if existing != nil {
			// Increment quantity on existing user cart item
			newQty := existing.Quantity + item.Quantity
			if err := s.repo.UpdateCartItemQuantity(ctx, existing.ID, newQty); err != nil {
				return nil, apperror.NewInternal("failed to merge item quantity", err)
			}
			// Delete guest item (already merged)
			if err := s.repo.DeleteCartItem(ctx, item.ID, guestCart.ID); err != nil {
				return nil, apperror.NewInternal("failed to remove merged guest item", err)
			}
		} else {
			// Move item to user cart
			if err := s.repo.MoveCartItem(ctx, item.ID, userCart.ID); err != nil {
				return nil, apperror.NewInternal("failed to move item to user cart", err)
			}
		}
	}

	// Mark guest cart as merged
	if err := s.repo.UpdateCartStatus(ctx, guestCart.ID, "merged"); err != nil {
		return nil, apperror.NewInternal("failed to mark guest cart as merged", err)
	}

	resp, err := s.repo.GetCartWithItems(ctx, userCart.ID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch merged cart", err)
	}
	return resp, nil
}

// findOrCreateCart locates an active cart or creates one.
func (s *cartService) findOrCreateCart(ctx context.Context, userID, sessionID string) (*Cart, error) {
	cart, err := s.repo.FindActiveCart(ctx, userID, sessionID)
	if err != nil {
		return nil, apperror.NewInternal("failed to find cart", err)
	}
	if cart != nil {
		return cart, nil
	}

	if sessionID == "" {
		return nil, apperror.NewInvalidInput("session ID required", nil)
	}

	cart, err = s.repo.CreateCart(ctx, userID, sessionID)
	if err != nil {
		return nil, apperror.NewInternal("failed to create cart", err)
	}
	return cart, nil
}
