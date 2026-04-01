-- Phase 3, PR 3A: Composite indexes, active cart uniqueness, order soft delete, shipping address validation

-- Composite indexes for common query patterns
CREATE INDEX idx_carts_user_status ON carts(user_id, status);
CREATE INDEX idx_carts_session_status ON carts(session_id, status);
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
CREATE INDEX idx_products_category_status ON products(category_id, status);

-- Ensure only one active cart per authenticated user
CREATE UNIQUE INDEX idx_one_active_cart_per_user
    ON carts(user_id)
    WHERE status = 'active' AND user_id IS NOT NULL;

-- Ensure only one active cart per guest session
CREATE UNIQUE INDEX idx_one_active_cart_per_session
    ON carts(session_id)
    WHERE status = 'active' AND user_id IS NULL;

-- Soft delete for orders (regulatory compliance — orders should never be hard-deleted)
ALTER TABLE orders ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at) WHERE deleted_at IS NULL;

-- Update RLS policy to exclude soft-deleted orders from user queries
DROP POLICY IF EXISTS "Users can view own orders" ON orders;
CREATE POLICY "Users can view own orders" ON orders
    FOR SELECT USING (auth.uid() = user_id AND deleted_at IS NULL);

-- Shipping address CHECK constraint — validate required JSONB keys
ALTER TABLE orders ADD CONSTRAINT shipping_address_required_fields
    CHECK (
        shipping_address ? 'name' AND
        shipping_address ? 'line1' AND
        shipping_address ? 'city' AND
        shipping_address ? 'zip' AND
        shipping_address ? 'country'
    );

-- Enforce numeric precision on cart_items.unit_price (was generic NUMERIC)
ALTER TABLE cart_items ALTER COLUMN unit_price TYPE numeric(10,2);

-- Guest cart RLS policy (for completeness if anon access is ever enabled)
CREATE POLICY "Guest users can view own carts by session" ON carts
    FOR SELECT USING (user_id IS NULL AND session_id IS NOT NULL);

-- Page views: add timestamp index for analytics queries
CREATE INDEX idx_page_views_viewed_at ON page_views(viewed_at DESC);
CREATE INDEX idx_page_views_product_time ON page_views(product_id, viewed_at DESC);
