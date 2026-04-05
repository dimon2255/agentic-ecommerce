# ADR-0002: Cart and Order Items Store Price Snapshots

## Status

Accepted

## Context

Product prices can change at any time (sales, repricing, corrections). When a customer adds an item to their cart or places an order, we need to decide whether to store the price at that moment or always look it up from the current product/SKU record.

This affects cart display accuracy, order totals, and dispute resolution.

## Options Considered

### Option A — Always fetch current price from product/SKU table
- Pro: Cart always shows latest price
- Con: Price can change between add-to-cart and checkout, surprising the customer
- Con: Historical orders would show wrong amounts if price changes retroactively

### Option B — Snapshot `unit_price` at creation time on cart items and order items
- Pro: Cart price is stable from add to checkout
- Pro: Order records are historically accurate — what the customer actually paid
- Pro: Simplifies dispute resolution and refund calculations
- Con: Cart can show a stale price if item sits in cart for days

## Decision

Use **Option B**. Both `cart_items.unit_price` and `order_items.unit_price` store the price at the time the item was added/ordered. The price is captured from the SKU record at insertion time.

## Consequences

- **Positive:** Order history is always accurate — no retroactive price distortion
- **Positive:** Checkout totals match what the customer saw when they added items
- **Negative:** Long-lived cart items may show outdated prices; a future enhancement could refresh prices at checkout time
- **Risks:** If a price drops, customers with stale cart snapshots pay the old higher price unless we add a refresh mechanism
