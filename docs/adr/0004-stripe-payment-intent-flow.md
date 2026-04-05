# ADR-0004: Stripe PaymentIntent with Client-Side Confirmation

## Status

Accepted

## Context

We need to integrate Stripe for payment processing. Stripe offers multiple integration patterns: Checkout Sessions (hosted page), PaymentIntent API (custom UI), and Payment Links (no-code). We need a flow that supports 3D Secure (SCA), gives us control over the checkout UX, and keeps sensitive card data off our server.

## Options Considered

### Option A — Stripe Checkout Sessions (hosted payment page)
- Pro: Stripe handles the entire payment UI; minimal frontend code
- Pro: Automatic 3D Secure, Apple Pay, Google Pay support
- Con: Redirects customer away from our site; limited UX customization
- Con: Harder to integrate with our existing cart/checkout flow

### Option B — PaymentIntent API with client-side confirmation
- Pro: Full control over checkout UX — payment form stays on our site
- Pro: Supports 3D Secure via Stripe.js `confirmPayment`
- Pro: Server creates PaymentIntent, client confirms — card data never hits our API
- Con: More frontend code to write (Stripe Elements integration)
- Con: Must handle more edge cases (payment failures, 3D Secure redirects)

### Option C — Payment Links (no-code)
- Pro: Zero implementation effort
- Con: No programmatic control; can't tie to our order/cart system

## Decision

Use **Option B**. The Go API creates a PaymentIntent (with order ID in metadata) and returns the `client_secret`. The Nuxt frontend uses Stripe.js to confirm payment client-side. Stripe webhooks notify the API of payment success/failure to update order status.

## Consequences

- **Positive:** Seamless checkout UX — customer never leaves the site
- **Positive:** 3D Secure/SCA handled automatically by Stripe.js
- **Positive:** Card data never touches our server (PCI scope minimized to SAQ-A)
- **Negative:** More frontend complexity managing Stripe Elements and payment states
- **Risks:** Webhook delivery failures could leave orders in limbo; webhook handler must be idempotent and verify signatures
