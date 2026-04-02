-- Phase 3, PR 3B: Checkout idempotency key to prevent duplicate orders
ALTER TABLE orders ADD COLUMN idempotency_key TEXT UNIQUE;
