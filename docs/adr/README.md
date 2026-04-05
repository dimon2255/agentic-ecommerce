# Architectural Decision Records

Lightweight docs capturing key technical decisions, their context, and trade-offs.
See [TEMPLATE.md](TEMPLATE.md) for the format.

| # | Decision | Status | Date |
|---|----------|--------|------|
| [0001](0001-service-role-key-pattern.md) | Go API bypasses RLS via service role key | Accepted | 2025-03 |
| [0002](0002-price-snapshot-strategy.md) | Cart/order items store unit_price at creation time | Accepted | 2025-03 |
| [0003](0003-ssr-routing-rules.md) | SSR for catalog pages, SPA for cart/checkout/auth | Accepted | 2025-03 |
| [0004](0004-stripe-payment-intent-flow.md) | Stripe PaymentIntent with client-side confirmation | Accepted | 2025-03 |
| [0005](0005-cart-session-identity.md) | Guest carts use X-Session-ID, auth carts use JWT user ID | Accepted | 2025-03 |
| [0006](0006-vitest-frontend-testing.md) | Vitest + @nuxt/test-utils for frontend test coverage | Accepted | 2026-04 |
