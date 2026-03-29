package cart

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CartHandler struct {
	db *supabase.Client
}

func NewCartHandler(db *supabase.Client) *CartHandler {
	return &CartHandler{db: db}
}

func (h *CartHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetCart)
	r.Post("/items", h.AddItem)
	r.Patch("/items/{itemId}", h.UpdateItem)
	r.Delete("/items/{itemId}", h.RemoveItem)
	r.Post("/merge", h.MergeCart)
	return r
}

// findActiveCart looks up the active cart for the current user or session.
func (h *CartHandler) findActiveCart(r *http.Request) *Cart {
	userID, hasUser := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")

	var carts []Cart
	q := h.db.From("carts").Select("*").Eq("status", "active")
	if hasUser && userID != "" {
		q = q.Eq("user_id", userID)
	} else if sessionID != "" {
		q = q.Eq("session_id", sessionID).Is("user_id", "null")
	} else {
		return nil
	}

	if err := q.Limit(1).Execute(&carts); err != nil || len(carts) == 0 {
		return nil
	}
	return &carts[0]
}

// findOrCreateCart returns the active cart or creates one.
func (h *CartHandler) findOrCreateCart(r *http.Request) (*Cart, error) {
	if cart := h.findActiveCart(r); cart != nil {
		return cart, nil
	}

	userID, _ := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		return nil, fmt.Errorf("session ID required")
	}

	newCart := map[string]any{
		"session_id": sessionID,
		"status":     "active",
	}
	if userID != "" {
		newCart["user_id"] = userID
	}

	var created []Cart
	if err := h.db.From("carts").Insert(newCart).Execute(&created); err != nil {
		return nil, fmt.Errorf("create cart: %w", err)
	}
	if len(created) == 0 {
		return nil, fmt.Errorf("cart not returned after creation")
	}
	return &created[0], nil
}

// getCartResponse fetches the full cart with enriched items.
func (h *CartHandler) getCartResponse(cartID string) (*CartResponse, error) {
	var items []CartItemWithSKU
	err := h.db.From("cart_items").
		Select("*,skus(sku_code,price_override,products(name,slug,base_price,images))").
		Eq("cart_id", cartID).
		Execute(&items)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []CartItemWithSKU{}
	}
	return &CartResponse{ID: cartID, Items: items}, nil
}

// lookupSKUPrice fetches the current price for a SKU.
func (h *CartHandler) lookupSKUPrice(skuID string) (float64, error) {
	var skus []SKUForPrice
	err := h.db.From("skus").
		Select("price_override,products(base_price)").
		Eq("id", skuID).
		Execute(&skus)
	if err != nil {
		return 0, err
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

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, hasUser := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")
	if (!hasUser || userID == "") && sessionID == "" {
		response.Error(w, http.StatusBadRequest, "authentication or session ID required")
		return
	}

	cart := h.findActiveCart(r)
	if cart == nil {
		response.JSON(w, http.StatusOK, CartResponse{Items: []CartItemWithSKU{}})
		return
	}

	resp, err := h.getCartResponse(cart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch cart items")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SKUID == "" || req.Quantity < 1 || req.Quantity > 999 {
		response.Error(w, http.StatusBadRequest, "sku_id is required and quantity must be between 1 and 999")
		return
	}

	cart, err := h.findOrCreateCart(r)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get or create cart")
		return
	}

	unitPrice, err := h.lookupSKUPrice(req.SKUID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid SKU")
		return
	}

	// Check for existing item with same SKU
	var existing []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("cart_id", cart.ID).Eq("sku_id", req.SKUID).
		Execute(&existing); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to check existing items")
		return
	}

	if len(existing) > 0 {
		// Increment quantity
		newQty := existing[0].Quantity + req.Quantity
		var updated []CartItem
		if err := h.db.From("cart_items").
			Update(map[string]any{"quantity": newQty}).
			Eq("id", existing[0].ID).
			Execute(&updated); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to update item quantity")
			return
		}
	} else {
		// Insert new item
		var inserted []CartItem
		if err := h.db.From("cart_items").Insert(map[string]any{
			"cart_id":    cart.ID,
			"sku_id":     req.SKUID,
			"quantity":   req.Quantity,
			"unit_price": unitPrice,
		}).Execute(&inserted); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to add item to cart")
			return
		}
	}

	resp, err := h.getCartResponse(cart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch updated cart")
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Quantity < 1 || req.Quantity > 999 {
		response.Error(w, http.StatusBadRequest, "quantity must be between 1 and 999")
		return
	}

	cart := h.findActiveCart(r)
	if cart == nil {
		response.Error(w, http.StatusNotFound, "cart not found")
		return
	}

	// Verify item belongs to this cart
	var items []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("id", itemID).Eq("cart_id", cart.ID).
		Execute(&items); err != nil || len(items) == 0 {
		response.Error(w, http.StatusNotFound, "cart item not found")
		return
	}

	var updated []CartItem
	if err := h.db.From("cart_items").
		Update(map[string]any{"quantity": req.Quantity}).
		Eq("id", itemID).
		Eq("cart_id", cart.ID).
		Execute(&updated); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update item")
		return
	}

	resp, err := h.getCartResponse(cart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch updated cart")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")

	cart := h.findActiveCart(r)
	if cart == nil {
		response.Error(w, http.StatusNotFound, "cart not found")
		return
	}

	// Verify item belongs to this cart
	var items []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("id", itemID).Eq("cart_id", cart.ID).
		Execute(&items); err != nil || len(items) == 0 {
		response.Error(w, http.StatusNotFound, "cart item not found")
		return
	}

	if err := h.db.From("cart_items").Delete().
		Eq("id", itemID).
		Eq("cart_id", cart.ID).
		Execute(nil); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to remove item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) MergeCart(w http.ResponseWriter, r *http.Request) {
	userID, hasUser := middleware.GetUserID(r.Context())
	if !hasUser || userID == "" {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req MergeCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SessionID == "" {
		response.Error(w, http.StatusBadRequest, "session_id is required")
		return
	}

	// Find guest cart
	var guestCarts []Cart
	if err := h.db.From("carts").Select("*").
		Eq("session_id", req.SessionID).
		Is("user_id", "null").
		Eq("status", "active").
		Execute(&guestCarts); err != nil || len(guestCarts) == 0 {
		// No guest cart to merge — return current user cart
		userCart := h.findUserCart(userID)
		if userCart == nil {
			response.JSON(w, http.StatusOK, CartResponse{Items: []CartItemWithSKU{}})
			return
		}
		resp, err := h.getCartResponse(userCart.ID)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to fetch cart")
			return
		}
		response.JSON(w, http.StatusOK, resp)
		return
	}
	guestCart := guestCarts[0]

	// Find or create user cart
	userCart := h.findUserCart(userID)
	if userCart == nil {
		var created []Cart
		if err := h.db.From("carts").Insert(map[string]any{
			"user_id":    userID,
			"session_id": req.SessionID,
			"status":     "active",
		}).Execute(&created); err != nil || len(created) == 0 {
			response.Error(w, http.StatusInternalServerError, "failed to create user cart")
			return
		}
		userCart = &created[0]
	}

	// Fetch guest cart items
	var guestItems []CartItem
	if err := h.db.From("cart_items").Select("*").
		Eq("cart_id", guestCart.ID).
		Execute(&guestItems); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch guest cart items")
		return
	}

	// Move each guest item to user cart
	for _, item := range guestItems {
		// Check for duplicate SKU in user cart
		var existing []CartItem
		if err := h.db.From("cart_items").Select("*").
			Eq("cart_id", userCart.ID).Eq("sku_id", item.SKUID).
			Execute(&existing); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to check duplicate items")
			return
		}

		if len(existing) > 0 {
			// Increment quantity on existing user cart item
			newQty := existing[0].Quantity + item.Quantity
			if err := h.db.From("cart_items").
				Update(map[string]any{"quantity": newQty}).
				Eq("id", existing[0].ID).
				Execute(nil); err != nil {
				response.Error(w, http.StatusInternalServerError, "failed to merge item quantity")
				return
			}
			// Delete guest item
			if err := h.db.From("cart_items").Delete().
				Eq("id", item.ID).Execute(nil); err != nil {
				response.Error(w, http.StatusInternalServerError, "failed to remove merged guest item")
				return
			}
		} else {
			// Move item to user cart
			if err := h.db.From("cart_items").
				Update(map[string]any{"cart_id": userCart.ID}).
				Eq("id", item.ID).
				Execute(nil); err != nil {
				response.Error(w, http.StatusInternalServerError, "failed to move item to user cart")
				return
			}
		}
	}

	// Mark guest cart as merged
	if err := h.db.From("carts").
		Update(map[string]any{"status": "merged"}).
		Eq("id", guestCart.ID).
		Execute(nil); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to mark guest cart as merged")
		return
	}

	resp, err := h.getCartResponse(userCart.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch merged cart")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// findUserCart looks up the active cart for a specific user ID.
func (h *CartHandler) findUserCart(userID string) *Cart {
	var carts []Cart
	if err := h.db.From("carts").Select("*").
		Eq("user_id", userID).
		Eq("status", "active").
		Limit(1).
		Execute(&carts); err != nil || len(carts) == 0 {
		return nil
	}
	return &carts[0]
}
