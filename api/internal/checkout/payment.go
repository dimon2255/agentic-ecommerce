package checkout

// PaymentService abstracts payment provider operations for testability.
type PaymentService interface {
	// CreatePaymentIntent creates a payment intent and returns the client secret and payment intent ID.
	CreatePaymentIntent(amountCents int64, currency, orderID string) (clientSecret, paymentIntentID string, err error)
	// VerifyWebhook verifies a webhook signature and returns the event type and payment intent ID.
	VerifyWebhook(payload []byte, sigHeader string) (eventType, paymentIntentID string, err error)
}
