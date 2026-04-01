package checkout

import (
	"context"
	"math"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
)

type checkoutService struct {
	repo            Repository
	payments        PaymentService
	paymentCurrency string
}

func NewService(repo Repository, payments PaymentService, paymentCurrency string) Service {
	return &checkoutService{
		repo:            repo,
		payments:        payments,
		paymentCurrency: paymentCurrency,
	}
}

func (s *checkoutService) StartCheckout(ctx context.Context, userID, sessionID string, req StartCheckoutRequest) (*StartCheckoutResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	activeCart, err := s.repo.FindActiveCart(ctx, userID, sessionID)
	if err != nil {
		return nil, apperror.NewInternal("failed to find cart", err)
	}
	if activeCart == nil {
		return nil, apperror.NewInvalidInput("no active cart found", nil)
	}

	items, err := s.repo.GetCartItems(ctx, activeCart.ID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch cart items", err)
	}
	if len(items) == 0 {
		return nil, apperror.NewInvalidInput("cart is empty", nil)
	}

	// Check for price changes
	var priceChanges []PriceChange
	for _, item := range items {
		currentPrice := item.SKU.Product.BasePrice
		if item.SKU.PriceOverride != nil {
			currentPrice = *item.SKU.PriceOverride
		}
		if currentPrice != item.UnitPrice {
			priceChanges = append(priceChanges, PriceChange{
				SKUID:    item.SKUID,
				SKUCode:  item.SKU.SKUCode,
				OldPrice: item.UnitPrice,
				NewPrice: currentPrice,
			})
			s.repo.UpdateCartItemPrice(ctx, item.ID, currentPrice)
		}
	}
	if len(priceChanges) > 0 {
		return nil, apperror.NewConflict("prices have changed", map[string]any{
			"price_changes": priceChanges,
		})
	}

	// Calculate totals
	var subtotal float64
	for _, item := range items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}
	total := subtotal

	// Create order
	orderData := map[string]any{
		"email":            req.Email,
		"shipping_address": req.ShippingAddress,
		"subtotal":         subtotal,
		"total":            total,
		"status":           "draft",
	}
	if userID != "" {
		orderData["user_id"] = userID
	}

	order, err := s.repo.CreateOrder(ctx, orderData)
	if err != nil || order == nil {
		return nil, apperror.NewInternal("failed to create order", err)
	}

	// Create order items
	for _, item := range items {
		orderItemData := map[string]any{
			"order_id":     order.ID,
			"sku_id":       item.SKUID,
			"product_name": item.SKU.Product.Name,
			"sku_code":     item.SKU.SKUCode,
			"quantity":     item.Quantity,
			"unit_price":   item.UnitPrice,
		}
		if err := s.repo.CreateOrderItem(ctx, orderItemData); err != nil {
			s.repo.DeleteOrder(ctx, order.ID)
			return nil, apperror.NewInternal("failed to create order items", err)
		}
	}

	// Create payment intent
	amountCents := int64(math.Round(total * 100))
	clientSecret, piID, err := s.payments.CreatePaymentIntent(amountCents, s.paymentCurrency, order.ID)
	if err != nil {
		s.repo.DeleteOrder(ctx, order.ID)
		return nil, apperror.NewInternal("failed to create payment intent", err)
	}

	// Update order with payment intent
	s.repo.UpdateOrder(ctx, order.ID, map[string]any{
		"stripe_payment_intent_id": piID,
		"status":                   "pending",
	})

	return &StartCheckoutResponse{
		OrderID:      order.ID,
		ClientSecret: clientSecret,
	}, nil
}

func (s *checkoutService) GetOrder(ctx context.Context, orderID string) (*OrderResponse, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch order", err)
	}
	if order == nil {
		return nil, apperror.NewNotFound("order")
	}

	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch order items", err)
	}

	itemResponses := make([]OrderItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = OrderItemResponse{
			ProductName: item.ProductName,
			SKUCode:     item.SKUCode,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	return &OrderResponse{
		ID:              order.ID,
		Status:          order.Status,
		Email:           order.Email,
		ShippingAddress: order.ShippingAddress,
		Subtotal:        order.Subtotal,
		Total:           order.Total,
		Items:           itemResponses,
		CreatedAt:       order.CreatedAt,
	}, nil
}

func (s *checkoutService) HandlePaymentSucceeded(ctx context.Context, piID string) error {
	s.repo.UpdateOrderByPaymentIntent(ctx, piID, map[string]any{"status": "paid"})

	order, _ := s.repo.FindOrderByPaymentIntent(ctx, piID)
	if order != nil && order.UserID != nil {
		s.repo.ExpireUserCarts(ctx, *order.UserID)
	}
	return nil
}

func (s *checkoutService) HandlePaymentFailed(ctx context.Context, piID string) error {
	s.repo.UpdateOrderByPaymentIntent(ctx, piID, map[string]any{"status": "cancelled"})
	return nil
}
