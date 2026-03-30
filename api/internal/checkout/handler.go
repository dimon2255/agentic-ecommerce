package checkout

import (
	"encoding/json"
	"io"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CheckoutHandler struct {
	db       *supabase.Client
	payments PaymentService
}

func NewCheckoutHandler(db *supabase.Client, payments PaymentService) *CheckoutHandler {
	return &CheckoutHandler{db: db, payments: payments}
}

func (h *CheckoutHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/start", h.StartCheckout)
	return r
}

func (h *CheckoutHandler) OrderRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", h.GetOrder)
	return r
}

func (h *CheckoutHandler) WebhookRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.HandleWebhook)
	return r
}

func (h *CheckoutHandler) StartCheckout(w http.ResponseWriter, r *http.Request) {
	var req StartCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		response.Error(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.ShippingAddress.Name == "" || req.ShippingAddress.Line1 == "" ||
		req.ShippingAddress.City == "" || req.ShippingAddress.Zip == "" ||
		req.ShippingAddress.Country == "" {
		response.Error(w, http.StatusBadRequest, "shipping address fields required: name, line1, city, zip, country")
		return
	}

	activeCart := h.findActiveCart(r)
	if activeCart == nil {
		response.Error(w, http.StatusBadRequest, "no active cart found")
		return
	}

	var items []cartItem
	err := h.db.From("cart_items").
		Select("*,skus(sku_code,price_override,products(name,base_price))").
		Eq("cart_id", activeCart.ID).
		Execute(&items)
	if err != nil || len(items) == 0 {
		response.Error(w, http.StatusBadRequest, "cart is empty")
		return
	}

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
			h.db.From("cart_items").
				Update(map[string]any{"unit_price": currentPrice}).
				Eq("id", item.ID).
				Execute(nil)
		}
	}
	if len(priceChanges) > 0 {
		response.JSON(w, http.StatusConflict, map[string]any{
			"error":         "prices have changed",
			"price_changes": priceChanges,
		})
		return
	}

	var subtotal float64
	for _, item := range items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}
	total := subtotal

	userID, _ := middleware.GetUserID(r.Context())
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

	var orders []Order
	if err := h.db.From("orders").Insert(orderData).Execute(&orders); err != nil || len(orders) == 0 {
		response.Error(w, http.StatusInternalServerError, "failed to create order")
		return
	}
	order := orders[0]

	for _, item := range items {
		orderItemData := map[string]any{
			"order_id":     order.ID,
			"sku_id":       item.SKUID,
			"product_name": item.SKU.Product.Name,
			"sku_code":     item.SKU.SKUCode,
			"quantity":     item.Quantity,
			"unit_price":   item.UnitPrice,
		}
		if err := h.db.From("order_items").Insert(orderItemData).Execute(nil); err != nil {
			h.db.From("orders").Delete().Eq("id", order.ID).Execute(nil)
			response.Error(w, http.StatusInternalServerError, "failed to create order items")
			return
		}
	}

	amountCents := int64(math.Round(total * 100))
	clientSecret, piID, err := h.payments.CreatePaymentIntent(amountCents, "usd", order.ID)
	if err != nil {
		h.db.From("orders").Delete().Eq("id", order.ID).Execute(nil)
		response.Error(w, http.StatusInternalServerError, "failed to create payment intent")
		return
	}

	h.db.From("orders").
		Update(map[string]any{
			"stripe_payment_intent_id": piID,
			"status":                   "pending",
		}).
		Eq("id", order.ID).
		Execute(nil)

	response.JSON(w, http.StatusOK, StartCheckoutResponse{
		OrderID:      order.ID,
		ClientSecret: clientSecret,
	})
}

func (h *CheckoutHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	var orders []Order
	err := h.db.From("orders").Select("*").Eq("id", orderID).Limit(1).Execute(&orders)
	if err != nil || len(orders) == 0 {
		response.Error(w, http.StatusNotFound, "order not found")
		return
	}
	order := orders[0]

	var items []OrderItem
	h.db.From("order_items").Select("*").Eq("order_id", orderID).Execute(&items)
	if items == nil {
		items = []OrderItem{}
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

	response.JSON(w, http.StatusOK, OrderResponse{
		ID:              order.ID,
		Status:          order.Status,
		Email:           order.Email,
		ShippingAddress: order.ShippingAddress,
		Subtotal:        order.Subtotal,
		Total:           order.Total,
		Items:           itemResponses,
		CreatedAt:       order.CreatedAt,
	})
}

func (h *CheckoutHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 65536))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	eventType, piID, err := h.payments.VerifyWebhook(payload, sigHeader)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid webhook signature")
		return
	}

	switch eventType {
	case "payment_intent.succeeded":
		h.db.From("orders").
			Update(map[string]any{"status": "paid"}).
			Eq("stripe_payment_intent_id", piID).
			Execute(nil)

		var orders []Order
		h.db.From("orders").Select("user_id").Eq("stripe_payment_intent_id", piID).Execute(&orders)
		if len(orders) > 0 && orders[0].UserID != nil {
			h.db.From("carts").
				Update(map[string]any{"status": "expired"}).
				Eq("user_id", *orders[0].UserID).
				Eq("status", "active").
				Execute(nil)
		}

	case "payment_intent.payment_failed":
		h.db.From("orders").
			Update(map[string]any{"status": "cancelled"}).
			Eq("stripe_payment_intent_id", piID).
			Execute(nil)
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *CheckoutHandler) findActiveCart(r *http.Request) *cart {
	userID, ok := middleware.GetUserID(r.Context())
	if ok {
		var carts []cart
		h.db.From("carts").Select("*").Eq("user_id", userID).Eq("status", "active").Limit(1).Execute(&carts)
		if len(carts) > 0 {
			return &carts[0]
		}
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID != "" {
		var carts []cart
		h.db.From("carts").Select("*").Eq("session_id", sessionID).Eq("status", "active").Is("user_id", "null").Limit(1).Execute(&carts)
		if len(carts) > 0 {
			return &carts[0]
		}
	}

	return nil
}
