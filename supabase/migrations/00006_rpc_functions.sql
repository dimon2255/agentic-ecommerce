-- Phase 3, PR 3B: Atomic cart item add/increment (eliminates race condition)
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
