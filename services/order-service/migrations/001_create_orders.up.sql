CREATE TYPE order_status AS ENUM (
    'PENDING',
    'CONFIRMED',
    'PAYMENT_FAILED',
    'OUT_OF_STOCK'
);

-- Order header. Line items live in order_items (an order may contain many).
CREATE TABLE orders (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id   VARCHAR(64) NOT NULL,
    total_amount  NUMERIC(10, 2) NOT NULL,
    status        order_status NOT NULL DEFAULT 'PENDING',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status   ON orders(status);

-- One row per item in an order. unit_price is captured at order time from the
-- catalog so historical orders are unaffected by later price changes.
CREATE TABLE order_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id     UUID NOT NULL,
    quantity    INT NOT NULL CHECK (quantity > 0),
    unit_price  NUMERIC(10, 2) NOT NULL,
    -- An item appears at most once per order (quantities are merged upstream).
    CONSTRAINT uq_order_items_order_item UNIQUE (order_id, item_id)
);

CREATE INDEX idx_order_items_order ON order_items(order_id);
