package checkout

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type CheckoutHandler struct {
	svc                Service
	payments           PaymentService
	webhookMaxBodySize int64
}

func NewCheckoutHandler(svc Service, payments PaymentService, webhookMaxBodySize int64) *CheckoutHandler {
	return &CheckoutHandler{
		svc:                svc,
		payments:           payments,
		webhookMaxBodySize: webhookMaxBodySize,
	}
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

	userID, _ := middleware.GetUserID(r.Context())
	sessionID := r.Header.Get("X-Session-ID")

	resp, err := h.svc.StartCheckout(r.Context(), userID, sessionID, req)
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *CheckoutHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetOrder(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.ErrorFromAppError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *CheckoutHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, h.webhookMaxBodySize))
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
		h.svc.HandlePaymentSucceeded(r.Context(), piID)
	case "payment_intent.payment_failed":
		h.svc.HandlePaymentFailed(r.Context(), piID)
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
