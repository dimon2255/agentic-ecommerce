-- Create order status enum
CREATE TYPE order_status AS ENUM ('draft', 'pending', 'paid', 'shipped', 'completed', 'cancelled');

-- Orders table
CREATE TABLE orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES auth.users(id) ON DELETE SET NULL,
    status order_status NOT NULL DEFAULT 'draft',
    email text NOT NULL,
    shipping_address jsonb NOT NULL,
    subtotal numeric(10,2) NOT NULL CHECK (subtotal >= 0),
    total numeric(10,2) NOT NULL CHECK (total >= 0),
    stripe_payment_intent_id text UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Order items table (all fields snapshotted at checkout time)
CREATE TABLE order_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sku_id uuid NOT NULL REFERENCES skus(id) ON DELETE RESTRICT,
    product_name text NOT NULL,
    sku_code text NOT NULL,
    quantity integer NOT NULL CHECK (quantity > 0),
    unit_price numeric(10,2) NOT NULL CHECK (unit_price >= 0)
);

-- Indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_stripe_pi ON orders(stripe_payment_intent_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- RLS
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own orders" ON orders
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Service role full access on orders" ON orders
    FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "Users can view own order items" ON order_items
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM orders WHERE orders.id = order_items.order_id AND orders.user_id = auth.uid())
    );

CREATE POLICY "Service role full access on order_items" ON order_items
    FOR ALL USING (auth.role() = 'service_role');

-- Trigger for updated_at
CREATE TRIGGER orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
