-- Phase 3, PR 3B: Atomic operations via RPC functions

-- Atomic cart item add/increment (eliminates race condition in check-then-insert)
CREATE OR REPLACE FUNCTION add_or_increment_cart_item(
    p_cart_id UUID,
    p_sku_id UUID,
    p_quantity INTEGER,
    p_unit_price NUMERIC(10,2)
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO cart_items (cart_id, sku_id, quantity, unit_price)
    VALUES (p_cart_id, p_sku_id, p_quantity, p_unit_price)
    ON CONFLICT (cart_id, sku_id)
    DO UPDATE SET
        quantity = cart_items.quantity + EXCLUDED.quantity,
        updated_at = now();
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Atomic order creation with items (eliminates checkout race condition)
CREATE OR REPLACE FUNCTION create_order_atomic(
    p_email TEXT,
    p_shipping_address JSONB,
    p_subtotal NUMERIC(10,2),
    p_total NUMERIC(10,2),
    p_user_id UUID DEFAULT NULL,
    p_items JSONB DEFAULT '[]'::JSONB
)
RETURNS UUID AS $$
DECLARE
    v_order_id UUID;
    v_item JSONB;
BEGIN
    -- Create the order
    INSERT INTO orders (email, shipping_address, subtotal, total, status, user_id)
    VALUES (p_email, p_shipping_address, p_subtotal, p_total, 'draft', p_user_id)
    RETURNING id INTO v_order_id;

    -- Create order items from the JSONB array
    FOR v_item IN SELECT * FROM jsonb_array_elements(p_items)
    LOOP
        INSERT INTO order_items (order_id, sku_id, product_name, sku_code, quantity, unit_price)
        VALUES (
            v_order_id,
            (v_item->>'sku_id')::UUID,
            v_item->>'product_name',
            v_item->>'sku_code',
            (v_item->>'quantity')::INTEGER,
            (v_item->>'unit_price')::NUMERIC(10,2)
        );
    END LOOP;

    RETURN v_order_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Checkout idempotency key to prevent duplicate orders
ALTER TABLE orders ADD COLUMN idempotency_key TEXT UNIQUE;
CREATE INDEX idx_orders_idempotency ON orders(idempotency_key) WHERE idempotency_key IS NOT NULL;
