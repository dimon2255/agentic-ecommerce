package stripe

import (
	"encoding/json"
	"fmt"

	gostripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
)

// Client wraps the Stripe API. Implements checkout.PaymentService.
type Client struct {
	webhookSecret string
}

// NewClient sets the Stripe API key globally and returns a client.
func NewClient(secretKey, webhookSecret string) *Client {
	gostripe.Key = secretKey
	return &Client{webhookSecret: webhookSecret}
}

func (c *Client) CreatePaymentIntent(amountCents int64, currency, orderID string) (string, string, error) {
	params := &gostripe.PaymentIntentParams{
		Amount:             gostripe.Int64(amountCents),
		Currency:           gostripe.String(currency),
		PaymentMethodTypes: gostripe.StringSlice([]string{"card"}),
	}
	params.AddMetadata("order_id", orderID)

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", "", fmt.Errorf("create payment intent: %w", err)
	}
	return pi.ClientSecret, pi.ID, nil
}

func (c *Client) VerifyWebhook(payload []byte, sigHeader string) (string, string, error) {
	event, err := webhook.ConstructEvent(payload, sigHeader, c.webhookSecret)
	if err != nil {
		return "", "", fmt.Errorf("verify webhook signature: %w", err)
	}

	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(event.Data.Raw, &data); err != nil {
		return string(event.Type), "", nil
	}
	return string(event.Type), data.ID, nil
}
