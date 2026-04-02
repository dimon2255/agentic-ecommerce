-- Phase 3, PR 3B: Atomic order creation with items (eliminates checkout race condition)
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
    INSERT INTO orders (email, shipping_address, subtotal, total, status, user_id)
    VALUES (p_email, p_shipping_address, p_subtotal, p_total, 'draft', p_user_id)
    RETURNING id INTO v_order_id;

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
