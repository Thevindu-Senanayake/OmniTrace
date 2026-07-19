CREATE TYPE order_status AS ENUM (
    'PENDING',
    'CONFIRMED',
    'PAYMENT_FAILED',
    'OUT_OF_STOCK'
);

CREATE TABLE orders (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id   VARCHAR(64) NOT NULL,
    item_id       UUID NOT NULL,
    quantity      INT NOT NULL CHECK (quantity > 0),
    total_amount  NUMERIC(10, 2) NOT NULL,
    status        order_status NOT NULL DEFAULT 'PENDING',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status   ON orders(status);
