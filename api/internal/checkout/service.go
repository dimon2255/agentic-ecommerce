package checkout

import "context"

// Service defines business operations for the checkout domain.
type Service interface {
	StartCheckout(ctx context.Context, userID, sessionID string, req StartCheckoutRequest) (*StartCheckoutResponse, error)
	GetOrder(ctx context.Context, orderID string) (*OrderResponse, error)
	HandlePaymentSucceeded(ctx context.Context, piID string) error
	HandlePaymentFailed(ctx context.Context, piID string) error
}
